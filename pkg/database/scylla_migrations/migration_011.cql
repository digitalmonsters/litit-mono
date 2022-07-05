create materialized view if not exists notification_all_with_push_admin as
select user_id, event_type, created_at, entity_id, related_entity_id
from notification
where user_id is not null
  and event_type in ('comment_reply', 'comment_vote_like', 'comment_vote_dislike', 'comment_profile_resource_create',
                     'comment_content_resource_create', 'follow', 'content_posted', 'tip', 'content_like',
                     'bonus_followers', 'bonus_time', 'content_upload', 'spot_upload', 'content_reject',
                     'kyc_status_verified', 'kyc_status_rejected', 'creator_status_rejected', 'creator_status_approved',
                     'creator_status_pending', 'first_daily_followers_bonus', 'first_daily_time_bonus',
                     'first_guest_x_earned_points', 'first_guest_x_paid_views', 'first_x_paid_views',
                     'first_referral_joined', 'first_video_shared', 'first_weekly_followers_bonus', 'first_weekly_time_bonus',
                     'first_x_paid_views_as_content_owner', 'guest_max_earned_points_for_views', 'increase_reward_stage_1',
                     'increase_reward_stage_2', 'registration_verify_bonus', 'other_referrals_joined', 'custom_reward_increase',
                     'megabonus', 'first_time_avatar_added', 'first_video_uploaded', 'first_spot_uploaded', 'add_description_bonus',
                     'first_x_paid_views_gender_push', 'first_email_marketing_added', 'top_daily_spot_bonus', 'top_weekly_spot_bonus',
                     'last_boring_spots', 'first_boring_spots', 'warning_boring_spots', 'push_admin')
  and entity_id is not null
  and related_entity_id is not null
  and created_at is not null
    PRIMARY KEY ((user_id), created_at, event_type, entity_id, related_entity_id)
with clustering order by (created_at desc);

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