---
title: "Entity"
linkTitle: "Entity"
date: 2021-11-20
description: >

---


```bash
+---------------------+
|       Entity        |
+---------------------+
|      Actions        |
|      State          |
|    Attributes       |
|     Settings        |
|      Metrics        |
|      Storage        |
+---------------------+
```
Это простая схема, которая представляет объект "Entity". 
Каждый компонент имеет свою роль и функциональность в управлении и мониторинге объекта "Entity" в умном доме.

Объект "Entity" является центральным элементом системы Smart Home и объединяет различные аспекты объекта, его состояния, 
атрибутов, действий, настроек и метрик. Это позволяет системе эффективно управлять и контролировать объекты в умном доме.

Вот подробное описание компонентов объекта "Entity":

1. Actions (Действия): Объект "Entity" может принимать и обрабатывать различные действия или команды. Действия представляют 
собой операции, которые можно выполнить над объектом, такие как включение, выключение, изменение параметров и т.д.

2. State (Состояние): Объект "Entity" имеет список состояний, в которые он может переходить. В каждый конкретный момент 
времени объект может находиться только в одном состоянии. Примеры состояний могут включать "Включено", "Выключено", 
"Режим ожидания", "Воспроизведение" и т.д.

3. Attributes (Атрибуты): Атрибуты представляют собой хранилище состояния объекта. Это заранее определенный перечень 
полей и свойств, который содержит информацию о текущем состоянии объекта. Атрибуты могут быть представлены в виде map[string]any, 
где ключ - имя атрибута, а значение - соответствующее значение атрибута.

4. Settings (Настройки): Настройки объекта "Entity" представляют собой неизменяемый заранее определенный перечень полей и свойств.
Они определяют конфигурационные параметры объекта, которые могут быть установлены во время его настройки. Настройки также могут 
быть представлены в виде map[string]any.

5. Metrics (Метрики): Метрики объекта "Entity" представляют собой информацию о его атрибутах или состоянии, которая используется 
для мониторинга и измерения производительности или поведения объекта. Метрики могут включать такие данные, как среднее значение 
атрибута, количество изменений состояния и т.д.

6. Storage (Хранилище): Хранилище объекта "Entity" предоставляет историю изменений его состояния или атрибутов. Оно записывает 
и сохраняет предыдущие значения, позволяя отслеживать и анализировать историю изменений объекта. Хранилище может использоваться 
для отображения графиков, аналитики или выполнения других операций с историческими данными объекта.

Объект "Entity" собирает все эти компоненты вместе, обеспечивая унифицированный и гибкий подход
к управлению и мониторингу различных устройств и систем в умном доме.

В системе Smart Home каждый объект "Entity" реализован на основе определенного плагина. Плагины предоставляют различные
функциональности и возможности для объектов "Entity" в системе. Некоторые популярные плагины, которые могут быть 
использованы для создания объектов "Entity", включают плагины sensor, mqtt, weather, automation, и другие.

Когда создается новый объект "Entity", он связывается с определенным плагином, который определяет его функциональность и
возможности. Например, если объект "Entity" представляет собой датчик, то для его реализации может использоваться плагин 
sensor. Если объект "Entity" предназначен для взаимодействия с брокером MQTT, то для него может быть использован плагин mqtt.

Каждый плагин предоставляет свои собственные хэндлеры (обработчики) и методы, которые позволяют объектам "Entity" выполнять 
определенные действия, получать данные, отправлять сообщения и т.д.

```bash
              +-----------------------+
              |                       |
              |     Smart Home        |
              |                       |
              +-----------------------+
                         |
                         |
                         v
              +-----------------------+
              |                       |
              |       Entities        |
              |                       |
              +-----------------------+
                         |
                         |
                         v
   +----------------------------------------+
   |                                        |
   |              Plugins                   |
   |                                        |
   +----------------------------------------+
   |             |           |              |
   |             |           |              |
   v             v           v              v
+----------+ +----------+ +----------+ +----------+
|          | |          | |          | |          |
|  Plugin  | |  Plugin  | |  Plugin  | |  Plugin  |
|          | |          | |          | |          |
+----------+ +----------+ +----------+ +----------+
   |             |           |              |
   |             |           |              |
   v             v           v              v
+----------+ +----------+ +----------+ +----------+
|          | |          | |          | |          |
| Entity 1 | | Entity 2 | | Entity 3 | | Entity 4 |
|          | |          | |          | |          |
+----------+ +----------+ +----------+ +----------+

                     ^
                     |
                     |
              +------------------+
              |                  |
              |  Automation      |
              |                  |
              +------------------+
```

На схеме представлена общая структура связи между объектами "Entity", плагинами и компонентом автоматизации в системе Smart Home.

1. Smart Home является центральной частью системы и координирует взаимодействие между всеми компонентами.
2. Объекты "Entity" представляют конкретные устройства, датчики или другие компоненты в системе. Каждый объект "Entity" может использовать определенный плагин для своей реализации.
3. Плагины предоставляют функциональность и возможности для объектов "Entity". Они содержат логику и методы, которые позволяют объектам "Entity" взаимодействовать с внешними устройствами, собирать данные, отправлять команды и т.д.
4. Компонент автоматизации отвечает за создание сценариев и запуск определенных действий на основе условий и триггеров. Он может использовать объекты "Entity" и их плагины для определения условий, триггеров и действий в сценариях.

Такая структура позволяет системе Smart Home быть гибкой и расширяемой, так как новые плагины и объекты "Entity" могут быть добавлены, 
а компонент автоматизации может использовать их для создания разнообразных сценариев и автоматизации действий.