set schema 'public';

create extension if not exists "uuid-ossp";

create table if not exists nodes
(
    id uuid default uuid_generate_v4() not null
    constraint nodes_pk primary key,
    addr text not null
);

create unique index if not exists nodes_addr_uindex on nodes (addr);



create table if not exists files
(
    id uuid default uuid_generate_v4() not null
    constraint files_pk primary key,
    location text not null
);

create unique index if not exists files_location_uindex on files (lower(location));

create table if not exists shards
(
    file_id uuid not null
    constraint shards_files_id_fk references files on delete cascade,
    node_id uuid not null
    constraint shards_nodes_id_fk references nodes,
    index smallint not null,
    size bigint not null,
    created_at timestamp not null default current_timestamp,
    status int not null default 1
);

create unique index if not exists shards_uindex on shards (file_id, node_id, index);



select n.id, n.addr, coalesce(sum(size), 0) ts
from nodes n left join shards s on n.id = s.node_id
group by (n.id, n.addr)
order by coalesce(sum(size), 0);
