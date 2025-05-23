# Тестовое задание для Cloud.ru Camp


## 1. Вопросы для разогрева

1. Опишите самую интересную задачу в программировании, которую вам приходилось решать?

- Самая интересная задача на моей памяти мне попалась на первом курсе ВУЗа. У нас тогда был предмет "Основы программирования на С++", несмотря на название - он был достаточно полным и предлагал очень много интересных задач, например poolAllocator совместимый с std или помехоустойчивый архиватор с кодами Хемминга. Но самое интересное на мой взгляд - интепретатор SQL запросов. На тот момент мы не знали что такое SQL, как работают Реляционный СУБД и что такое паттерны проектирования. По итогу получилось изучить и SQL, и бд, и написать in-memory базу данных с хоть и ограниченным, но довольно большим набором команд.


2. Расскажите о своем самом большом факапе? Что вы предприняли для решения проблемы?

-  Как то мы работали над групповым проектом по одной из дисциплин в университете. Мы писали приложение на C#, это было RESTApi с подключением к бд PostgreSQL. Хоть это был и учебный проект, мы имитировали высокие нагрузки и поэтому в нашей бд было довольно много записей. Я всегда стремился оптимизировать свой код, выбирать самые быстрые варианты из доступных и выжимать максимум из доступных ресурсов. Я решил что наши запросы могли бы работать быстрее, поэтому предложил добавить индексы в таблицы бд. Я тогда еще не знал что такое уровни изоляции, что создание индекса это дорогая и впервую очередь блокирующая операция, поэтому когда я написал CREATE INDEX наша база встала, а до сдачи проекта оставалось совсем немного. Пришлось насильно остановить бд и отказаться от оптимизаций запросов, но после этого я изучил вопрос создания индекса в высоконагруженных системах, зачем нужен CONCURRENTLY, помимо этого улучшил и сам индекс чтобы добиться index only scan в плане запроса и ускорить итоговую производительность базы на два порядка.

3. Каковы ваши ожидания от участия в буткемпе?

- Я ожидаю в первую очередь развить свои способности, причем наверное soft-скиллы у меня в приоритете. Стажировка для меня это всегда не просто работа, это в первую очередь возможности, нетворкинг, общение с людьми из айти-сферы. Жду интересные задачи, я прочитал о компании Cloud.ru и понимаю, что здесь наверняка много уникальных проблем и решений, к тому же некоторые мои коллеги по учебе, которые в ней работали, рассказывают очень много положительного. Также очень хотелось бы продолжить свою карьеру вместе с этой компанией и получить оффер.

---

## 2. Разработка HTTP-балансировщика нагрузки на Go
### Часть 1. Балансировщик нагрузки
**Основной функционал:**
    Модуль балансировщика представляет функционал балансировщика нагрузки по стратегии round robin.
    Для настройки приложения нужно создать конфигурационный файл в .json формате и передать его как входной аргумент при запуске.
    Для корректной работы также нужно запустить сервисы, на которые будет перераспределяться нагрузка. Для этого можно воспользоваться кодом из [main.go](/main.go) и запустить его в отдельном терминале или прописать в [конфиге](/cmd/balancer/config.json) адреса собственных серверов.
    
Запустить можно с помощью
```bash
make test_lb
```


**Распределение запросов:**
    Многопоточная работа приложения достигается с помощью внутренних механизмов GO для работы с http-запросами. Для сохранности данных и избежания datarace, ресурсы защищены мьютексами и другими примитивами синхронизации из пакета sync.

**Логирование:**
    В приложении поддерживаются несколько уровней и форматов логирования, логгер основан на log/slog из стандартной библиотеки. 
    Если у меня было больше времени, я бы добавил больше логов для большей ясности во время работы приложения и в идеале добавил бы интеграцию с ELK или fluentD для аккумулирования логов и дальнейшей обработки.

**Конфигурация:**
Конфигурация происходит через файл config.json, путь до которого передается через аргументы командной строки. 
В текущей версии конфигурации можно прописать:
- порт приложения, на который нужно присылать запросы 
- различные таймауты 
- конфигурация текущей среды разработки(prod/test...)
- список url на которые нужно балансировать запросы.

**Что бы я добавил**
Если бы у меня было больше времени, я бы добавил:
- Docker и docker-compose.yaml для деплоя приложения
- Интеграционные и mock тесты, для проверки крайних случаев и общего функционала
- Больше логов
- Дополнительные стратегии балансировки (weight rr и пр...)
- Интеграцию с ELK для мониторинга логов
- Интеграцию с Prometheus и Grafana для анализа метрик


### Часть 2. Реализация Rate-Limiting
**Реализация алгоритма Token Bucket:**
В моей реализации, бакеты хранятся в in-memory хранилище на основе map, но в идеале я бы реализовал внешнее хранилище на Redis или другой быстрой key-value базе данных. Также стоит посмотреть на бенчмарки и сравнить обычную map с sync.Map, вероятно в нашем юзкейсе она может оказаться быстрее. 
У каждого бакета есть свои значения rps, на основании которых высчитывается оптимальный таймер для time.Ticker по которому обновляются значения бакетов. На каждый бакет существует своя горутина, которая и обновляет токены.

**API**
Поддерживаются два вида запросов: 
- POST запрос для добавления конфигураций
- Запросы на сервер, который стоит за rate-limiter. 
Все запросы, кроме POST проходят через rate-limiter и проходят дальше\отбрасываются в зависимости от количества токенов отправителя.

**Гранулярное ограничение:**
С помощью api можно добавить конфигурацию клиента через POST запрос, вот пример:
```bash
curl -H 'Content-Type: application/json' \ 
-d '{ "ip":"192.168.160.1","capacity":10, "rate_per_sec": 2}'  \ 
-X POST    localhost:3000/config
``` 
POST http-запрос конфигурации отдельных пользователей (пользователи идентифицируются своим ip адресом). При добавлении такой конфигурации, бакет, закрепленный за пользователем, обновится и начнет считать токены по обновленным данным.
При старте приложения из базы данных достаются уже существующие конфигурации и на основании них создаются изначальные бакеты. При поступлении запроса от нового пользователя, для него автоматически создается свой бакет.

**Конкурентность:**
Потокобезопасность достигается с помощью механизмов пакета http и примитивов синхронизации пакета sync. Каждый бакет обновляется своей горутиной, на случай обновления конфигурации есть мьютексы как на каждом бакете, так и на всем хранилище бакетов? чтобы избежать гонок данных.
Для конкуретного доступа к бд используются транзакции и пул соединений, чтобы разграничивать запросы при асинхронном доступе к строкам таблицы.

**Хранилище**
Для хранения конфигураций выбрана СУБД PostgreSQL.
Написаны миграции для базы, которые автоматически применяются при старте приложения. Добавлен индекс для оптимизации поиска.


**Конфигурация**
В этом приложении можно сконфигурировать:
- окружение
- URL таргета
- Порт приложения
- Default конфигурации бакетов для новых пользователей
- Ограничения на подключения к бд
Также, через переменные окружения нужно определить параметры для подключения к БД (Например, через .env с дальнейшим использованием в docker-compose.yaml)

**Запуск**
Поднять контейнер в Docker можно командой:
```bash
docker compose up
```
или (предпочтительно)
```Makefile
make deploy_rate
```
После этого, можно сделать запрос:
```bash
curl http://localhost:3000
```
и должно вывести вернуть {"message":"HelloWorld"} из контейнера на 8080 порту.
**Тесты**

При тестировании через Apache Bench с командой 
```bash
ab -n 5000 -c 1000 http://localhost:8080/
```
Конфигурация:
```json
"user_config": {
    "tokens": 1000,
    "rate_per_sec": 1000
  }
```
Результат(Вцелом, похоже на правду):
```bash
Server Hostname:        localhost
Server Port:            3000

Document Path:          /
Document Length:        24 bytes

Concurrency Level:      1000
Time taken for tests:   2.140 seconds
Complete requests:      5000
Failed requests:        2072
   (Connect: 0, Receive: 0, Length: 2072, Exceptions: 0)
Non-2xx responses:      2072
Total transferred:      555408 bytes
HTML transferred:       70272 bytes
Requests per second:    2336.69 [#/sec] (mean)
Time per request:       427.956 [ms] (mean)
Time per request:       0.428 [ms] (mean, across all concurrent requests)
Transfer rate:          253.48 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0   58 555.5      6    6770
Processing:   -23  995 1964.3    168    8914
Waiting:        0  918 1874.4    243    8904
Total:          0 1053 2018.6    276    8947

Percentage of the requests served within a certain time (ms)
  50%    276
  66%    575
  75%   1055
  80%   1260
  90%   1919
  95%   6820
  98%   7132
  99%   7347
 100%   8947 (longest request)
```

**Что бы я добавил**
Если бы у меня было больше времени, я бы добавил:
- Интеграционные тесты и mock тесты, для проверки отдельных частей приложения и функционала вцелом
- Бенчмарки на разных реализациях хранилища бакетов
- Логи на всех уровнях приложения
- Интеграции? что и в первой части


## Что бы я сделал вцелом для проекта
- Github actions для авмоматических тестов и линтера
- Больше комментариев godoc
- Examples для демонстрации работоспособности и вариантов использования модулей
- Кастомные ошибки


