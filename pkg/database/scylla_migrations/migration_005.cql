create table if not exists user
(
    cluster_key         bigint,
    user_id             bigint,
    username            text,
    firstname           text,
    lastname            text,
    language            text,
    name_privacy_status int,
    email               text,
    primary key ((cluster_key), user_id)
) with default_time_to_live = 7776000;