create table if not exists notification
(
    user_id             bigint,
    event_type          text,
    entity_id           bigint,
    related_entity_id   bigint,
    created_at          timestamp,
    notifications_count bigint,
    title               text,
    body                text,
    headline            text,
    kind                text,
    rendering_variables text,
    custom_data         text,
    notification_info   text,
    primary key ((user_id, event_type), created_at, entity_id, related_entity_id)
) with clustering order by (created_at desc) and default_time_to_live = 7776000;

create table if not exists notification_relation
(
    user_id           bigint,
    event_type        text,
    entity_id         bigint,
    related_entity_id bigint,
    event_applied     boolean,
    primary key ((user_id, event_type), entity_id, related_entity_id)
) with default_time_to_live = 7776000;

create table if not exists push_notification_group_queue
(
    deadline_key       timestamp,
    deadline           timestamp,
    user_id            bigint,
    event_type         text,
    entity_id          bigint,
    created_at         timestamp,
    notification_count bigint,
    primary key ((deadline_key, deadline), user_id, event_type, entity_id)
) with default_time_to_live = 86400;