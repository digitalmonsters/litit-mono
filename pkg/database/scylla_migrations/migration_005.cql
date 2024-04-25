create materialized view if not exists notification_comment as
select user_id, event_type, created_at, entity_id, related_entity_id
from notification
where user_id is not null
  and event_type in ('comment_reply', 'comment_vote_like', 'comment_vote_dislike', 'comment_profile_resource_create',
                     'comment_content_resource_create')
  and entity_id is not null
  and related_entity_id is not null
  and created_at is not null
    PRIMARY KEY ((user_id), created_at, event_type, entity_id, related_entity_id)
with clustering order by (created_at desc);

create materialized view if not exists notification_following as
select user_id, event_type, created_at, entity_id, related_entity_id
from notification
where user_id is not null
  and event_type = 'follow'
  and entity_id is not null
  and related_entity_id is not null
  and created_at is not null
    PRIMARY KEY ((user_id), created_at, event_type, entity_id, related_entity_id)
with clustering order by (created_at desc);
