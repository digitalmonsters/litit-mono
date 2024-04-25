create materialized view if not exists notification_system_with_push_admin as
select user_id, event_type, created_at, entity_id, related_entity_id
from notification
where user_id is not null
  and event_type = 'push_admin'
  and entity_id is not null
  and related_entity_id is not null
  and created_at is not null
    PRIMARY KEY ((user_id), created_at, event_type, entity_id, related_entity_id)
with clustering order by (created_at desc);