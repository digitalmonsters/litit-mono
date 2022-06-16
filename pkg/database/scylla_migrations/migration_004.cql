create table user_notifications_read
(
    cluster_key         bigint,
    notification_id     bigint,
    user_id             bigint,
    primary key (cluster_key, notification_id, user_id)
);
