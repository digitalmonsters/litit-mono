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
	}
}
