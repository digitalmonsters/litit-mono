create table user_notifications_settings
(
    cluster_key bigint,
    user_id     bigint,
    template_id text,
    muted       boolean,
    primary key (cluster_key, user_id, template_id)
);
