
---
title: "Alexa"
linkTitle: "alexa"
date: 2021-10-20
description: >
  
---

{{< alert color="warning" >}}Не законченная статья{{< /alert >}}

В системе **Smart Home** реализован плагин Amazon Alexa, который обеспечивает интеграцию с голосовым помощником Amazon
Alexa. Этот плагин позволяет управлять устройствами и выполнить различные операции в вашем умном доме с помощью
голосовых команд.

Плагин Amazon Alexa обеспечивает следующий функционал:

1. Распознавание и обработка голосовых команд: Плагин позволяет принимать голосовые команды от пользователя,
   отправленные
2. через голосового помощника Amazon Alexa. Он распознает команды, анализирует их и выполняет соответствующие действия в
   вашем умном доме.

3. Управление устройствами: Плагин позволяет управлять устройствами в вашем умном доме с помощью голосовых команд.
4. Например, вы можете сказать "Alexa, включи свет в гостиной" или "Alexa, установи температуру на 25 градусов".

5. Интеграция с другими функциями системы: Плагин Amazon Alexa интегрируется с другими функциями системы **Smart Home**.
6. Например, вы можете использовать голосовые команды для запуска сценариев, управления автоматическими задачами или
7. получения информации о состоянии устройств.

Для использования плагина Amazon Alexa в вашем проекте **Smart Home** требуется настройка и подключение к вашему аккаунту
Amazon Alexa. После этого вы можете настроить голосовые команды и действия для управления вашим умным домом через
голосового помощника Amazon Alexa.

В плагине Amazon Alexa для системы **Smart Home** реализованы следующие хэндлеры на языке JavaScript:

1. `skillOnLaunch`: Этот хэндлер вызывается при запуске навыка (skill) на устройстве Amazon Alexa. Он предназначен для
2. обработки начального события запуска навыка. Вы можете определить здесь логику или выполнить необходимые действия при
3. запуске навыка. Пример использования:

```javascript
skillOnLaunch = () => {
    // Логика обработки события запуска навыка
};
```

2. `skillOnSessionEnd`: Этот хэндлер вызывается при завершении сессии с навыком. Он позволяет выполнить определенные
3. действия при окончании сеанса с пользователем. Например, вы можете сохранить состояние или выполнить завершающие
   операции. Пример использования:

```javascript
skillOnSessionEnd = () => {
    // Логика завершения сессии
};
```

3. `skillOnIntent`: Этот хэндлер вызывается при получении навыком интентов (намерений) от пользователя. Он предназначен
4. для обработки различных интентов и выполнения соответствующих действий. Вы можете определить здесь логику обработки
5. интентов и взаимодействия с вашим умным домом. Пример использования:

```javascript
skillOnIntent = () => {
    // Логика обработки интентов
};
```

Эти хэндлеры предоставляют возможность обрабатывать события запуска навыка, завершения сессии и получения интентов от
пользователя в контексте плагина Amazon Alexa. Вы можете определить свою логику и выполнить необходимые действия в
каждом из этих хэндлеров в соответствии с требованиями вашего проекта.

### пример кода

```coffeescript
skillOnLaunch = ->
#print '---action onLaunch---'
  Done('skillOnLaunch')
skillOnSessionEnd = ->
#print '---action onSessionEnd---'
  Done('skillOnSessionEnd')
skillOnIntent = ->
#print '---action onIntent---'
  state = 'on'
  if Alexa.slots['state'] == 'off'
    state = 'off'

  place = Alexa.slots['place']

  Done("#{place}_#{state}")

  Alexa.sendMessage("#{place}_#{state}")
```