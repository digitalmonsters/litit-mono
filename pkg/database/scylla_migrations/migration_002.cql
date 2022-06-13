create table if not exists notification
(
    user_id             int,
    event_type          text,
    entity_id           int,
    related_entity_id   int,
    created_at          timestamp,
    notifications_count int,

    title               text,
    body                text,
    headline            text,
    rendering_variables text,
    custom_data         text,
    notification_info   text,
    primary key ((user_id, event_type), created_at, entity_id, related_entity_id)
) with clustering order by (created_at desc) and default_time_to_live = 7776000; -- 90 days

create table if not exists notification_relation
(
    user_id           int,
    event_type        text,
    entity_id         int,
    related_entity_id int,
    event_applied     boolean,
    primary key ((user_id, event_type), entity_id, related_entity_id)
) with default_time_to_live = 7776000; -- 90 days

create table if not exists push_notification_group_queue
(
    deadline_key       timestamp,
    deadline           timestamp,
    user_id            int,
    event_type         text,
    entity_id          int,
    created_at         timestamp,
    notification_count int,
    primary key ((deadline_key, deadline), user_id, event_type, entity_id)
) with default_time_to_live = 86400; -- 1 day
