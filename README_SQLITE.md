# CompareFK — вариант с SQLite storage

Что изменено:

- полностью убрано накопление `masterGuids`, `slaveGuids`, `resultGuids` в RAM
- master GUID складываются во временную SQLite БД
- совпадения из slave складываются в отдельную SQLite таблицу
- итоговый результат пишется потоково в файл, без сборки всего результата в памяти
- статистика стала потокобезопасной
- исправлены захваты переменных в goroutine
- чтение SQL-файлов сортируется по имени
- запись файлов идёт через buffered writer

## Новые параметры config.yml

```yaml
mode: difference
masterdsn: postgresql://user:pass@host:5432/db
slavedsn: postgresql://user:pass@host:5432/db
loglevel: info
limit: 50000
ratelimit: 4
mastersql: ./Master/
slavesql: ./Slave.sql
storage: sqlite
sqlite_path: compare.db
```

## Как работает

1. Каждый master SQL выполняется на master БД.
2. Все GUID из master пишутся в локальную SQLite таблицу `master_guids`.
3. Потом `master_guids` читается кусками по `limit`.
4. Для каждого куска выполняется `Slave.sql` на slave БД.
5. Найденные GUID пишутся в SQLite таблицу `matched_guids`.
6. Финальный результат:
   - `difference` = `master_guids - matched_guids`
   - `intersection` = `master_guids ∩ matched_guids`
7. Результат пишется потоково в `<mode>.txt`.

## Что это даёт

На больших объёмах память перестаёт расти пропорционально числу GUID. Основной объём уходит на диск, а в RAM живут только:

- активный chunk master GUID
- текущие rows/query objects
- буферы записи
- page cache SQLite

## Ограничения

Текущий `Slave.sql` всё ещё должен принимать chunk GUID как один параметр, как и в старой версии.
```go
slaveDB.QueryContext(ctx, slaveSQL, chunk)
```

## Практические стартовые настройки

Для 12+ млн строк:

- `limit: 50000`
- `ratelimit: 4`

Если slave-запрос тяжёлый, можно начать даже с:

- `limit: 20000`
- `ratelimit: 2`
