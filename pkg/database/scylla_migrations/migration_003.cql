create table user_notifications_read_counter
(
    notification_id     bigint,
    read_count          counter,
    primary key (notification_id)
);
