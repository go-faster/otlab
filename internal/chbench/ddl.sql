create table values
(
    id         UInt128,          -- hash(name)
    resource   UInt128,          -- hash(resource)   -> resources.id
    attributes UInt128,          -- hash(attributes) -> attributes.id
    timestamp  DateTime64(9),    -- ts
    value      Float64
)
    engine = MergeTree PARTITION BY toYearWeek(timestamp)
        ORDER BY (id, resource, attributes, timestamp)
        SETTINGS index_granularity = 8192;

create table metrics(name  String)
  engine = ReplacingMergeTree ORDER BY name;

--- estimated values: server_count * services_per_server
--- estimating as 10_000_000 (10M) for 100k servers with 100 services each.
create table resources
(
    id    UInt128, -- hash
    value String   -- value
)
engine = ReplacingMergeTree ORDER BY id;

create table attributes
(
    id    UInt128,  -- hash
    value String    -- value
)
engine = ReplacingMergeTree ORDER BY id;


create table resources_kv
(
    id    UInt128, -- hash

    -- map(key, value)
    keys   Array(String),
    values Array(String)
)
    engine = ReplacingMergeTree ORDER BY id;

-- CREATE VIEW meta.table_info AS
select
    parts.*,
    columns.compressed_size,
    columns.uncompressed_size,
    columns.ratio
from (
         select database,
                table,
                formatReadableSize(sum(data_uncompressed_bytes))          AS uncompressed_size,
                formatReadableSize(sum(data_compressed_bytes))            AS compressed_size,
                sum(data_compressed_bytes) / sum(data_uncompressed_bytes) AS ratio
         from system.columns
         group by database, table
         ) columns right join (
    select database,
           table,
           sum(rows)                                            as rows,
           max(modification_time)                               as latest_modification,
           formatReadableSize(sum(bytes))                       as disk_size,
           formatReadableSize(sum(primary_key_bytes_in_memory)) as primary_keys_size,
           any(engine)                                          as engine,
           sum(bytes)                                           as bytes_size
    from system.parts
    where active
    group by database, table
    ) parts on ( columns.database = parts.database and columns.table = parts.table )
order by parts.bytes_size desc;
