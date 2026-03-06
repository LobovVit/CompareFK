# CompareFK — SQLite-версия с единым конфигом

В проекте убрано хранение полного набора master/slave GUID в памяти. Вместо этого используется локальная SQLite-база и потоковая выгрузка результата.

## Что изменено

- единый `config.yml`
- пути к SQL лежат в конфиге:
  - `master_sql_dir`
  - `master_sql_glob`
  - `master_sql_files`
  - `slave_sql_file`
- оптимизации SQLite:
  - `WAL`
  - `synchronous=NORMAL`
  - `temp_store=FILE`
  - настраиваемый `cache_size`
  - `busy_timeout`
  - `mmap_size`
- результат пишется потоково в файл
- статистика и выполненные SQL сохраняются в папку запуска

## Структура

```text
config.yml
sql/
  master/
    0_master.sql
    1_master.sql
    2_master.sql
    3_master.sql
  slave/
    Slave.sql
runs/
```

## Запуск

```bash
go mod tidy
go build ./...
./compare -c ./config.yml
```

## Рекомендованный старт для 12+ млн GUID

```yaml
limit: 50000
ratelimit: 4
sqlite_write_batch: 20000
sqlite_cache_size_kb: 65536
sqlite_mmap_size_mb: 256
```

## Как указывать master SQL

### Каталог

```yaml
master_sql_dir: ./sql/master
```

### Glob-маска

```yaml
master_sql_glob: ./sql/master/*.sql
```

### Явный список файлов

```yaml
master_sql_files:
  - ./sql/master/0_master.sql
  - ./sql/master/1_master.sql
  - ./sql/master/2_master.sql
  - ./sql/master/3_master.sql
```

При наличии `master_sql_files` он имеет приоритет над `master_sql_glob` и `master_sql_dir`.

## Выходные файлы

После запуска создаётся папка `runs/YYYY_MM_DD_hh_mm_ss/`, внутри:

- `difference.txt` или `intersection.txt`
- `stat.txt`
- `sql/` — копии выполненных SQL
- `work/` — рабочая SQLite БД на время выполнения
