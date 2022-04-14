package database

import (
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func getMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "initial_node_tables_20220310",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"create table if not exists \"Devices\"\n(\n    id          uuid                     default gen_random_uuid() not null\n        primary key,\n    \"userId\"    integer                                            not null,\n    \"deviceId\"  varchar(255)                                       not null,\n    \"pushToken\" varchar(255)                                       not null,\n    platform    varchar(255)                                       not null,\n    \"createdAt\" timestamp with time zone default CURRENT_TIMESTAMP not null,\n    \"updatedAt\" timestamp with time zone,\n    \"deletedAt\" timestamp with time zone\n);\n\ncreate index if not exists \"Devices_user_idx\"\n    on \"Devices\" (\"userId\");\n\n",
					"create table if not exists notifications\n(\n    id                     uuid                     default gen_random_uuid() not null\n        primary key,\n    user_id                integer                                            not null,\n    type                   varchar(255)                                       not null,\n    title                  varchar(255)                                       not null,\n    message                varchar(255)                                       not null,\n    related_user_id        integer,\n    comment_id             integer,\n    comment                jsonb,\n    content_id             integer,\n    content                jsonb,\n    question_id            integer,\n    created_at             timestamp with time zone default CURRENT_TIMESTAMP not null,\n    kyc_reason             varchar(255),\n    kyc_status             varchar(255),\n    content_creator_status integer\n);\n\ncreate index if not exists notifications_search_idx\n    on notifications (user_id, type, created_at);\n\ncreate index if not exists notifications_user_idx\n    on notifications (user_id);\n\n",
					"create table if not exists \"SequelizeMeta\"\n(\n    name varchar(255) not null\n        primary key\n);",
					"create table if not exists \"Tasks\"\n(\n    id                 uuid                     default gen_random_uuid() not null\n        primary key,\n    priority           integer                                            not null,\n    status             integer                                            not null,\n    \"notificationType\" varchar(255)                                       not null,\n    metadata           jsonb,\n    \"createdAt\"        timestamp with time zone default CURRENT_TIMESTAMP not null,\n    \"updatedAt\"        timestamp with time zone\n);\n\ncomment on column \"Tasks\".priority is '0 - low, 1 - medium, 2 - high';\n\ncomment on column \"Tasks\".status is '0 - todo, 1 - in progress, 2 - success, 3 - failed, 4 - background';\n\ncomment on column \"Tasks\".\"notificationType\" is 'the type of notification which is uniq for each one';\n\ncreate index if not exists \"Tasks_createdAt_idx\"\n    on \"Tasks\" (\"createdAt\");\n\ncreate index if not exists \"Tasks_notificationType_idx\"\n    on \"Tasks\" (\"notificationType\");\n\ncreate index if not exists \"Tasks_priority_idx\"\n    on \"Tasks\" (priority);\n\ncreate index if not exists \"Tasks_status_idx\"\n    on \"Tasks\" (status);\n\n",
					"create table if not exists user_notifications\n(\n    user_id      integer           not null\n        primary key,\n    unread_count integer default 0 not null\n);\n\ncreate index if not exists user_notifications_unread_count_idx\n    on user_notifications (unread_count);\n\n",
				)
			},
		},
		{
			ID: "additional_tables_20220310",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"create table if not exists render_templates\n(\n    id         text      not null\n        constraint render_templates_pk\n            primary key,\n    title      text,\n    body       text,\n    created_at timestamp not null,\n    updated_at timestamp not null\n);",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('first_x_paid_views', 'You earned your first {{.total_earned}} LIT points for watching video ',\n        'You earn {{.view_bonus}} LIT point for each video you watched for {{.watch_threshold}}% of the time or more',\n        '2022-03-10 12:32:42.000000', '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('signup_bonus', 'You earned your first {{.reward_amount}} LIT points for joining Lit.it',\n        'You can invite friends and earn {{.invite_reward_amount}} LIT points for each friend who will join. ',\n        '2022-03-10 12:32:42.000000', '2022-03-10 12:32:42.000000');",
				)
			},
		},
		{
			ID: "additional_render_templates_20220310",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('first_referral_joined',\n        'Your first friend just joined Lit.it! You earned {{.referral_bonus}} LIT points.',\n        'You can share videos with your friends on Lit.it and earn {{.share_bonus}} LIT points for each share.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('increase_reward_stage_1',\n        'Limited time deal! {{.percent_multiplier}}% increased rewards for friends invitations.',\n        'Earn {{.new_referrer_verify_reward}} LIT points for each friend who joins via your link for the next {{.period_hours}} hours only.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('increase_reward_stage_2',\n        'Limited time deal! {{.percent_multiplier}}% increased rewards for friends invitations.',\n        'Earn {{.new_referrer_verify_reward}} LIT points for each friend who joins via your link for the next {{.period_hours}} hours only.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('first_video_shared',\n        'You shared your first video on Lit.it. You earned {{.share_bonus}} LIT points.',\n        'Did you know: if you share a video with a person that does not have Lit.it and he will join, you earn {{.referral_bonus}} LIT points for that join as well.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('first_x_paid_views_as_content_owner',\n        'Your videos just generated your first {{.views_content_owner_bonus_count}} views on Lit.it. You got {{.views_content_owner_bonus_total_earned}} LIT points.',\n        'You earn {{.views_content_owner_bonus}} LIT point for each view your videos generate. Keep posting amazing content and earn more LIT points from your video views.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('top_x_in_subcategory',\n        'You just got to the TOP {{.top_in_category_bonus_place}} creators in {{.category_name}} category on Lit.it. You got {{.top_in_category_bonus}} LIT points.',\n        'You earn {{.top_in_category_bonus}} LIT points each time when you stay in the TOP {{.top_in_category_bonus_place}} of category for the whole day.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('first_daily_time_bonus',\n        'You just got your daily bonus for “Time on app” +{{.daily_time_bonus}} LIT points.',\n        'If you do this daily bonus 7 days in a row you get a weekly super bonus {{.weekly_time_bonus}} LIT points.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('first_weekly_time_bonus',\n        'You just got your weekly bonus for “Time on app” +{{.weekly_time_bonus}} LIT points',\n        'You are doing great! You can keep accumulating weekly bonus points and transfer them to your points wallet anytime.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('first_daily_followers_bonus',\n        'You just got your daily bonus for “Followers gained” on lit.it +{{.daily_followers_bonus}} LIT points',\n        'If you do this daily bonus 7 days in a row you get a weekly super bonus of {{.weekly_followers_bonus}} LIT points.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('first_weekly_followers_bonus',\n        'You just got your weekly bonus for “Followers gained” on lit.it +{{.weekly_followers_bonus}} LIT points',\n        'You are doing great! You can keep accumulating weekly bonus points and transfer them to your points wallet anytime.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"DELETE FROM public.render_templates as rt WHERE rt.id = 'signup_bonus';",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('registration_verify_bonus',\n        'You earned your first {{.reward_amount}} LIT points for joining Lit.it',\n        'You can invite friends and earn {{.referrer_verify_reward_amount}} LIT points for each friend who will join.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
				)
			},
		},
		{
			ID: "update_render_templates_20220312",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"UPDATE public.render_templates SET title = 'You earned your first {{.verify_reward_amount}} LIT points for joining Lit.it' WHERE id = 'registration_verify_bonus';",
					"UPDATE public.render_templates SET title = ('Congratulations! ' || title) WHERE id in ('registration_verify_bonus', 'first_x_paid_views', 'first_referral_joined', 'first_video_shared', 'first_x_paid_views_as_content_owner', 'top_x_in_subcategory');",
				)
			},
		},
		{
			ID: "creator_status_templates_20220312",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('creator_status_rejected',\n        'Rejected creator approval.',\n        'Your Creator approval process has been rejected.',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at)\nVALUES ('creator_status_approved',\n        'Creator status approved.',\n         'Your Creator status has been approved',\n        '2022-03-10 12:32:42.000000',\n        '2022-03-10 12:32:42.000000');",
				)
			},
		},
		{
			ID: "add_kind_to_template_20220311",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"alter table render_templates add column if not exists kind text;",
					"update render_templates set kind = 'popup' where 1 = 1 ",
				)
			},
		},
		{
			ID: "guest_notifications_templates_150320221520",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('first_guest_x_paid_views', 'Create your account and get your first {{.signup_bonus}}  LIT points', 'On Lit.it you earn rewards for watching & sharing videos or inviting people', '2022-03-15 15:08:08.000000', '2022-03-15 15:08:10.000000', 'popup') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('first_guest_x_earned_points', 'You just earned your first {{.earned_points}} LIT points. Create your account to transfer them to your points wallet.', 'On Lit.it you earn rewards for watching & sharing videos or inviting people', '2022-03-15 15:16:35.000000', '2022-03-15 15:16:37.000000', 'popup') on conflict do nothing;")
			},
		},
		{
			ID: "add_headline_20220315",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db, "UPDATE public.render_templates\nSET title = 'You just got to the TOP {{.top_in_category_bonus_place}} creators in {{.category_name}} category on Lit.it. You got {{.top_in_category_bonus}} LIT points.'\nWHERE id LIKE 'top#_x#_in#_subcategory' ESCAPE '#';\n\nUPDATE public.render_templates\nSET title = 'Your first friend just joined Lit.it! You earned {{.referral_bonus}} LIT points.'\nWHERE id LIKE 'first#_referral#_joined' ESCAPE '#';\n\nUPDATE public.render_templates\nSET title = 'You shared your first video on Lit.it. You earned {{.share_bonus}} LIT points.'\nWHERE id LIKE 'first#_video#_shared' ESCAPE '#';\n\nUPDATE public.render_templates\nSET title = 'You earned your first {{.verify_reward_amount}} LIT points for joining Lit.it'\nWHERE id LIKE 'registration#_verify#_bonus' ESCAPE '#';\n\nUPDATE public.render_templates\nSET title = 'Your videos just generated your first {{.views_content_owner_bonus_count}} views on Lit.it. You got {{.views_content_owner_bonus_total_earned}} LIT points.'\nWHERE id LIKE 'first#_x#_paid#_views#_as#_content#_owner' ESCAPE '#';\n\nUPDATE public.render_templates\nSET title = 'You earned your first {{.total_earned}} LIT points for watching video '\nWHERE id LIKE 'first#_x#_paid#_views' ESCAPE '#';",
					"alter table render_templates add column if not exists headline text;",
					"update render_templates set headline = 'Congratulations!' where id in ('first_x_paid_views', 'first_referral_joined', 'first_video_shared', 'first_x_paid_views_as_content_owner', 'top_x_in_subcategory', 'registration_verify_bonus')",
					"UPDATE public.render_templates\nSET title = '{{.percent_multiplier}}% increased rewards for friends invitations.'\nWHERE id LIKE 'increase#_reward#_stage#_1' ESCAPE '#';\n\nUPDATE public.render_templates\nSET title = '{{.percent_multiplier}}% increased rewards for friends invitations.'\nWHERE id LIKE 'increase#_reward#_stage#_2' ESCAPE '#';",
					"update render_templates set headline = 'Limited time deal!' where id in ('increase_reward_stage_1', 'increase_reward_stage_2')",
				)
			},
		},
		{
			ID: "modify_headline_202203161316",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"update render_templates set headline = 'Congrats!' where id in ('first_x_paid_views', 'first_referral_joined', 'first_video_shared', 'first_x_paid_views_as_content_owner', 'top_x_in_subcategory', 'registration_verify_bonus', 'first_daily_time_bonus', 'first_daily_followers_bonus');",
					"update render_templates set headline = 'Get 3x more!' where id = 'increase_reward_stage_2';",
				)
			},
		},
		{
			ID: "change_creators_data_20220317",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"update render_templates set kind = 'content_creator' where id in ('creator_status_rejected','creator_status_approved')")
			},
		},
		{
			ID: "add_content_creator_pending_220317",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"INSERT INTO render_templates (id, title, body, created_at, updated_at, kind, headline) VALUES ('creator_status_pending', 'Creator status pending.', 'Your Creator approval process has been successfully initiated', '2022-03-17 13:40:04.000000', '2022-03-17 13:40:06.000000', 'content_creator', null) on conflict do nothing;",
				)
			},
		},
		{
			ID: "guest_max_earned_points_for_views_17032022",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind, headline) VALUES ('guest_max_earned_points_for_views'::text, 'You just earned maximum allowed LIT points as unregistered user.'::text, 'You can earn 100 LIT points for watching videos as unregistered user. Create your Lit.it account now to keep getting points for watching videos.'::text, '2022-03-17 18:35:38.000000'::timestamp, '2022-03-17 18:35:40.000000'::timestamp, 'popup'::text, 'Congrats!'::text) on conflict do nothing;",
					"UPDATE render_templates SET headline = 'Congrats!' WHERE id = 'first_guest_x_paid_views'",
					"UPDATE render_templates SET headline = 'Congrats!' WHERE id = 'first_guest_x_earned_points'",
				)
			},
		},
		{
			ID: "other_referrals_joined_template_21032022",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind, headline) VALUES ('other_referrals_joined'::text, "+
						"'Your friend {{.username}} just joined via your link. +{{.referral_bonus}} LIT points'::text, null, '2022-03-21 19:35:38.000000'::timestamp, "+
						"'2022-03-21 19:35:38.000000'::timestamp, 'popup'::text, 'Congrats!'::text) on conflict do nothing;",
				)
			},
		},
		{
			ID: "add_comments_templates_180320221600",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('comment_content_resource_create', '{{.firstname}} {{.lastname}} commented your video', '{{.firstname}} {{.lastname}} commented: {{.comment}}', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.content.comment') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('comment_profile_resource_create', '{{.firstname}} {{.lastname}} commented your profile', '{{.firstname}} {{.lastname}} commented: {{.comment}}', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.profile.comment') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('comment_reply', '{{.firstname}} {{.lastname}} replied on your comment', '{{.firstname}} {{.lastname}} replied on your comment: {{.comment}}', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.comment.reply') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('comment_vote_like', '{{.firstname}} {{.lastname}} liked your comment', '{{.firstname}} {{.lastname}} liked your comment: {{.comment}}', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.comment.vote') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('comment_vote_dislike', '{{.firstname}} {{.lastname}} disliked your comment', '{{.firstname}} {{.lastname}} disliked your comment: {{.comment}}', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.comment.vote') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('content_like', '{{.firstname}} {{.lastname}}', '{{.firstname}} {{.lastname}} liked your video', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.content.like') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('content_upload', 'Video uploaded', 'Your video was successfully uploaded', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.content.successful-upload') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('content_reject', 'Your video is rejected', 'You were rejected to publish your video due to {{.reason}} content', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.content.rejected') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('kyc_status_verified', 'Verification is approved', 'Your identity verification has been approved', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.kyc.status') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('kyc_status_rejected', 'Verification is rejected', 'Your identity verification has been rejected Reason: {{.reason}}', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.kyc.status') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('follow', '{{.firstname}} {{.lastname}}', '{{.firstname}} {{.lastname}} started following you', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.profile.following') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('tip', '{{.firstname}} {{.lastname}}', '{{.firstname}} {{.lastname}} tipped you {{.pointsAmount}} LIT points', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.tip') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('bonus_time', 'Daily bonus', 'Daily reward for views', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.bonus.time') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('bonus_followers', 'Daily bonus', 'You received daily bonus for followers', '2022-03-18 16:00:00.000000', '2022-03-18 16:00:00.000000', 'push.bonus.followers') on conflict do nothing;",
				)
			},
		},
		{
			ID: "add_devices_userid_deviceid_uindex_290320221650",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db, `
					DELETE FROM "Devices" T1
						USING   "Devices" T2
					WHERE   T1.ctid < T2.ctid
						AND T1."userId" = T2."userId"
						AND T1."deviceId"  = T2."deviceId";

					create unique index if not exists devices_userid_deviceid_uindex
						on "Devices" ("userId", "deviceId");
				`)
			},
		},
		{
			ID: "add_content_posted_template_310320221200",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('content_posted', '{{.firstname}} {{.lastname}}', '{{.firstname}} {{.lastname}} posted new content', '2022-03-31 12:00:00.000000', '2022-03-31 12:00:00.000000', 'push.content.new-posted') on conflict do nothing;",
				)
			},
		},
		{
			ID: "update_other_referrals_verified_310320221400",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"update public.render_templates set body='You can share videos with your friends on Lit.it and earn 10 LIT points for each share.' "+
						"where id = 'other_referrals_joined'")
			},
		},
		{
			ID: "update_bonus_time_kind_010420221326",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"update public.render_templates set kind='push.bonus.daily' where id = 'bonus_time'")
			},
		},
		{
			ID: "set_different_type_to_ref_popup_20220401",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"UPDATE public.render_templates SET kind = 'default' WHERE id LIKE 'other#_referrals#_joined' ESCAPE '#';")
			},
		},
		{
			ID: "add_megabonus_template_050420221756",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('megabonus', 'Congrats! Mega bonus earned!', 'You just earned {{.pointsAmount}} LIT points for {{.referralsTarget}} friends invited to Lit.it', '2022-04-05 17:56:00.000000', '2022-04-05 17:56:00.000000', 'push.referral.megabonus') on conflict do nothing;",
				)
			},
		},
		{
			ID: "add_rendering_data_20220407",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"alter table notifications add column if not exists rendering_variables jsonb default '{}' not null;")
			},
		},
		{
			ID: "add_after_install_signup_templates_140420221100",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('guest_after_install_first_push', 'Complete your account creation & start earning LIT rewards', '', '2022-04-14 11:00:00.000000', '2022-04-14 11:00:00.000000', 'push.guest.after_install') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('guest_after_install_second_push', 'On Lit.it the more viral videos you watch the more your earn', '', '2022-04-14 11:00:00.000000', '2022-04-14 11:00:00.000000', 'push.guest.after_install') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('guest_after_install_third_push', 'Check this out. We picked our TOP viral videos for you', '', '2022-04-14 11:00:00.000000', '2022-04-14 11:00:00.000000', 'push.guest.after_install') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('user_after_signup_first_push', 'Few things you need to know about Lit.it', '', '2022-04-14 11:00:00.000000', '2022-04-14 11:00:00.000000', 'push.user.after_signup') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('user_after_signup_second_push', 'Check who earned the most for inviting friends to Lit.it', '', '2022-04-14 11:00:00.000000', '2022-04-14 11:00:00.000000', 'push.user.after_signup') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('user_after_signup_third_push', 'Who earned the most LIT points? Check this out', '', '2022-04-14 11:00:00.000000', '2022-04-14 11:00:00.000000', 'push.user.after_signup') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('user_after_signup_fourth_push', 'Check out TOP viral videos on Lit.it & earn LIT points', '', '2022-04-14 11:00:00.000000', '2022-04-14 11:00:00.000000', 'push.user.after_signup') on conflict do nothing;",
					"INSERT INTO public.render_templates (id, title, body, created_at, updated_at, kind) VALUES ('user_after_signup_fifth_push', 'How many LIT points in your wallet? Check this out', '', '2022-04-14 11:00:00.000000', '2022-04-14 11:00:00.000000', 'push.user.after_signup') on conflict do nothing;",
				)
			},
		},
		{
			ID: "add_custom_data_140420221724",
			Migrate: func(db *gorm.DB) error {
				return boilerplate_testing.ExecutePostgresSql(db,
					"alter table notifications add column if not exists custom_data jsonb default '{}' not null;")
			},
		},
	}
}
