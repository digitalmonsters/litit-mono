module github.com/digitalmonsters/litit-mono

go 1.25.3

require github.com/go-chi/chi/v5 v5.2.3

replace github.com/digitalmonsters/go-common => ./internal/common

replace github.com/digitalmonsters/ads-manager => ./internal/ads

replace github.com/digitalmonsters/comments => ./internal/comments

replace github.com/digitalmonsters/notification-handler => ./internal/notifications

replace github.com/digitalmonsters/configurator => ./internal/configurator
