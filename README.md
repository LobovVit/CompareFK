# CompareFK

Переработанная версия проекта для больших объёмов данных.

Основной режим работы теперь — через локальную SQLite, чтобы резко снизить расход RAM при сравнении больших наборов GUID.

# CompareFK + web monitor

В проект добавлен простой встроенный web-сервер для онлайн-наблюдения за выполнением.

## Адреса

После запуска, если в `config.yml` указано:

```yaml
web_enabled: true
web_listen: ":8080"
```

страницы будут доступны так:

- HTML: `http://localhost:8080/`
- JSON: `http://localhost:8080/api/status`
- health: `http://localhost:8080/healthz`

## Поля конфига

```yaml
web_enabled: true
web_listen: ":8080"
web_refresh_sec: 2
web_read_timeout_sec: 5
web_write_timeout_sec: 30
```

## Что считается строками

- для `master` — количество GUID, считанных из master SQL и загруженных в SQLite
- для `slave_chunk` — количество GUID, найденных в slave для конкретного куска master
- для `z_compute_*` — количество GUID, выгруженных в финальный результат
- для `slave.sql` — количество GUID master, уже переданных в slave-проверку по чанкам

`slave.sql` здесь показывает прогресс обхода master-чанков, а не число найденных совпадений.


