// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2023, Filippov Alex
//
// This library is free software: you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Library General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public
// License along with this library.  If not, see
// <https://www.gnu.org/licenses/>.

package media

import (
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/rtmp"
	"github.com/deepch/vdk/format/rtspv2"
)

// StreamServerRunStreamDo stream run do mux
func StreamServerRunStreamDo(streamID string, channelID string) {
	var status int
	defer func() {
		//TODO fix it no need unlock run if delete stream
		if status != 2 {
			Storage.StreamChannelUnlock(streamID, channelID)
		}
	}()
	for {

		//log.Info("Run stream")
		opt, err := Storage.StreamChannelControl(streamID, channelID)
		if err != nil {
			//log.Info("Exit", err)
			return
		}
		if opt.OnDemand && !Storage.ClientHas(streamID, channelID) {
			//log.Info("Stop stream no client")
			return
		}
		status, err = StreamServerRunStream(streamID, channelID, opt)
		if status > 0 {
			//log.Info("Stream exit by signal or not client")
			return
		}
		if err != nil {
			//log.Error(err.Error())
		}
		time.Sleep(2 * time.Second)

	}
}

// StreamServerRunStream core stream
func StreamServerRunStream(streamID string, channelID string, opt *ChannelST) (int, error) {
	if url, err := url.Parse(opt.URL); err == nil && strings.ToLower(url.Scheme) == "rtmp" {
		return StreamServerRunStreamRTMP(streamID, channelID, opt)
	}
	keyTest := time.NewTimer(20 * time.Second)
	checkClients := time.NewTimer(20 * time.Second)
	var start bool
	var fps int
	var preKeyTS = time.Duration(0)
	var Seq []*av.Packet
	RTSPClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{URL: opt.URL, InsecureSkipVerify: opt.InsecureSkipVerify, DisableAudio: !opt.Audio, DialTimeout: 3 * time.Second, ReadWriteTimeout: 5 * time.Second, Debug: opt.Debug, OutgoingProxy: true})
	if err != nil {
		return 0, err
	}
	Storage.StreamChannelStatus(streamID, channelID, ONLINE)
	defer func() {
		RTSPClient.Close()
		Storage.StreamChannelStatus(streamID, channelID, OFFLINE)
		Storage.StreamHLSFlush(streamID, channelID)
	}()
	var WaitCodec bool
	/*
		Example wait codec
	*/
	if RTSPClient.WaitCodec {
		WaitCodec = true
	} else {
		if len(RTSPClient.CodecData) > 0 {
			Storage.StreamChannelCodecsUpdate(streamID, channelID, RTSPClient.CodecData, RTSPClient.SDPRaw)
		}
	}
	//log.Info("Success connection RTSP")
	var ProbeCount int
	var ProbeFrame int
	var ProbePTS time.Duration
	Storage.NewHLSMuxer(streamID, channelID)
	defer Storage.HLSMuxerClose(streamID, channelID)
	for {
		select {
		//Check stream have clients
		case <-checkClients.C:
			if opt.OnDemand && !Storage.ClientHas(streamID, channelID) {
				return 1, ErrorStreamNoClients
			}
			checkClients.Reset(20 * time.Second)
		//Check stream send key
		case <-keyTest.C:
			return 0, ErrorStreamNoVideo
		//Read core signals
		case signals := <-opt.signals:
			switch signals {
			case SignalStreamStop:
				return 2, ErrorStreamStopCoreSignal
			case SignalStreamRestart:
				return 0, ErrorStreamRestart
			case SignalStreamClient:
				return 1, ErrorStreamNoClients
			}
		//Read rtsp signals
		case signals := <-RTSPClient.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				Storage.StreamChannelCodecsUpdate(streamID, channelID, RTSPClient.CodecData, RTSPClient.SDPRaw)
				WaitCodec = false
			case rtspv2.SignalStreamRTPStop:
				return 0, ErrorStreamStopRTSPSignal
			}
		case packetRTP := <-RTSPClient.OutgoingProxyQueue:
			Storage.StreamChannelCastProxy(streamID, channelID, packetRTP)
		case packetAV := <-RTSPClient.OutgoingPacketQueue:
			if WaitCodec {
				continue
			}

			if packetAV.IsKeyFrame {
				keyTest.Reset(20 * time.Second)
				if preKeyTS > 0 {
					Storage.StreamHLSAdd(streamID, channelID, Seq, packetAV.Time-preKeyTS)
					Seq = []*av.Packet{}
				}
				preKeyTS = packetAV.Time
			}
			Seq = append(Seq, packetAV)
			Storage.StreamChannelCast(streamID, channelID, packetAV)
			/*
			   HLS LL Test
			*/
			if packetAV.IsKeyFrame && !start {
				start = true
			}
			/*
				FPS mode probe
			*/
			if start {
				ProbePTS += packetAV.Duration
				ProbeFrame++
				if packetAV.IsKeyFrame && ProbePTS.Seconds() >= 1 {
					ProbeCount++
					if ProbeCount == 2 {
						fps = int(math.Round(float64(ProbeFrame) / ProbePTS.Seconds()))
					}
					ProbeFrame = 0
					ProbePTS = 0
				}
			}
			if start && fps != 0 {
				//TODO fix it
				packetAV.Duration = time.Duration((float32(1000)/float32(fps))*1000*1000) * time.Nanosecond
				Storage.HlsMuxerSetFPS(streamID, channelID, fps)
				Storage.HlsMuxerWritePacket(streamID, channelID, packetAV)
			}
		}
	}
}
func StreamServerRunStreamRTMP(streamID string, channelID string, opt *ChannelST) (int, error) {
	keyTest := time.NewTimer(20 * time.Second)
	checkClients := time.NewTimer(20 * time.Second)
	OutgoingPacketQueue := make(chan *av.Packet, 1000)
	Signals := make(chan int, 100)
	var start bool
	var fps int
	var preKeyTS = time.Duration(0)
	var Seq []*av.Packet

	conn, err := rtmp.DialTimeout(opt.URL, 3*time.Second)
	if err != nil {
		return 0, err
	}

	Storage.StreamChannelStatus(streamID, channelID, ONLINE)
	defer func() {
		conn.Close()
		Storage.StreamChannelStatus(streamID, channelID, OFFLINE)
		Storage.StreamHLSFlush(streamID, channelID)
	}()
	var WaitCodec bool

	codecs, err := conn.Streams()
	if err != nil {
		return 0, err
	}
	preDur := make([]time.Duration, len(codecs))
	Storage.StreamChannelCodecsUpdate(streamID, channelID, codecs, []byte{})

	log.Info("Success connection RTSP")
	var ProbeCount int
	var ProbeFrame int
	var ProbePTS time.Duration
	Storage.NewHLSMuxer(streamID, channelID)
	defer Storage.HLSMuxerClose(streamID, channelID)

	go func() {
		for {
			ptk, err := conn.ReadPacket()
			if err != nil {
				break
			}
			OutgoingPacketQueue <- &ptk
		}
		Signals <- 1
	}()

	for {
		select {
		//Check stream have clients
		case <-checkClients.C:
			if opt.OnDemand && !Storage.ClientHas(streamID, channelID) {
				return 1, ErrorStreamNoClients
			}
			checkClients.Reset(20 * time.Second)
		//Check stream send key
		case <-keyTest.C:
			return 0, ErrorStreamNoVideo
		//Read core signals
		case signals := <-opt.signals:
			switch signals {
			case SignalStreamStop:
				return 2, ErrorStreamStopCoreSignal
			case SignalStreamRestart:
				return 0, ErrorStreamRestart
			case SignalStreamClient:
				return 1, ErrorStreamNoClients
			}
		//Read rtsp signals
		case <-Signals:
			return 0, ErrorStreamStopRTSPSignal
		case packetAV := <-OutgoingPacketQueue:
			if preDur[packetAV.Idx] != 0 {
				packetAV.Duration = packetAV.Time - preDur[packetAV.Idx]
			}

			preDur[packetAV.Idx] = packetAV.Time

			if WaitCodec {
				continue
			}

			if packetAV.IsKeyFrame {
				keyTest.Reset(20 * time.Second)
				if preKeyTS > 0 {
					Storage.StreamHLSAdd(streamID, channelID, Seq, packetAV.Time-preKeyTS)
					Seq = []*av.Packet{}
				}
				preKeyTS = packetAV.Time
			}
			Seq = append(Seq, packetAV)
			Storage.StreamChannelCast(streamID, channelID, packetAV)
			/*
			   HLS LL Test
			*/
			if packetAV.IsKeyFrame && !start {
				start = true
			}
			/*
				FPS mode probe
			*/
			if start {
				ProbePTS += packetAV.Duration
				ProbeFrame++
				if packetAV.IsKeyFrame && ProbePTS.Seconds() >= 1 {
					ProbeCount++
					if ProbeCount == 2 {
						fps = int(math.Round(float64(ProbeFrame) / ProbePTS.Seconds()))
					}
					ProbeFrame = 0
					ProbePTS = 0
				}
			}
			if start && fps != 0 {
				//TODO fix it
				packetAV.Duration = time.Duration((float32(1000)/float32(fps))*1000*1000) * time.Nanosecond
				Storage.HlsMuxerSetFPS(streamID, channelID, fps)
				Storage.HlsMuxerWritePacket(streamID, channelID, packetAV)
			}
		}
	}
}