
create table "aml_cross_chain" (
      "id" bigserial not null,
      "match_id" int8,
      "tx_from_address" varchar,
      "chain" varchar not null,
      "number" int8 not null,
      "ts" timestamp not null,
      "index" int8 not null,
      "hash" varchar not null,
      "match_hash" varchar,
      "action_id" int8 not null,
      "project" varchar not null,
      "contract" varchar not null,
      "direction" varchar not null,
      "from_chain_id" numeric(256),
      "from_address" varchar,
      "to_chain_id" numeric(256),
      "to_address" varchar,
      "token" varchar not null,
      "amount" numeric(256),
      "real_token_in" varchar,
      "real_amount_in" numeric(256),
      "real_token_out" varchar,
      "real_amount_out" numeric(256),
      "match_tag" varchar not null,
      "detail" json,
      primary key (
                   "project",
                   "chain",
                   "number",
                   "hash",
                   "index",
                   "action_id"
          )
)partition by list(chain);

alter table aml_cross_chain add check (chain is not null);
alter table aml_cross_chain add check (number is not null);
alter table aml_cross_chain add check (ts is not null);
alter table aml_cross_chain add check (hash is not null);
alter table aml_cross_chain add check (action_id is not null);
alter table aml_cross_chain add check (project is not null);
alter table aml_cross_chain add check (contract is not null);
alter table aml_cross_chain add check (direction is not null);
alter table aml_cross_chain add check (token is not null);
alter table aml_cross_chain add check (match_tag is not null);

create table aml_cross_chain_eth partition of aml_cross_chain for values in ('eth');

create table aml_cross_chain_bsc partition of aml_cross_chain for values in ('bsc');

create table aml_cross_chain_polygon partition of aml_cross_chain for values in ('polygon');

create table aml_cross_chain_fantom partition of aml_cross_chain for values in ('fantom');

create table aml_cross_chain_avalanche partition of aml_cross_chain for values in ('avalanche');

create table aml_cross_chain_optimism partition of aml_cross_chain for values in ('optimism');

create table aml_cross_chain_arbitrum partition of aml_cross_chain for values in ('arbitrum');

create table aml_cross_chain_default partition of aml_cross_chain default;

create index on aml_cross_chain (id);

create index on aml_cross_chain (hash, match_hash);

create index on aml_cross_chain (tx_from_address);

create index on aml_cross_chain (match_id);

create index on aml_cross_chain (chain, project, match_tag);

create index on aml_cross_chain (from_address);

create index on aml_cross_chain (to_address);

create index on aml_cross_chain (chain, from_address);

create index on aml_cross_chain (chain, to_address);

create index on aml_cross_chain (ts desc);

create index on aml_cross_chain (project, match_id, real_token_in, real_token_out);

create index on aml_cross_chain (match_tag, project)