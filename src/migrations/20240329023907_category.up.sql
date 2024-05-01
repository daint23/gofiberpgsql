create table public."category" (
  id serial not null,
  name character varying(100) not null,
  description character varying(100),
  primary key(id)
)