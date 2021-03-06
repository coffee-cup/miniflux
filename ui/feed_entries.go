// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui  // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/model"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

// ShowFeedEntries shows all entries for the given feed.
func (c *Controller) ShowFeedEntries(w http.ResponseWriter, r *http.Request) {
	user, err := c.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, err)
		return
	}

	feedID, err := request.IntParam(r, "feedID")
	if err != nil {
		html.BadRequest(w, err)
		return
	}

	feed, err := c.store.FeedByID(user.ID, feedID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	if feed == nil {
		html.NotFound(w)
		return
	}

	offset := request.QueryIntParam(r, "offset", 0)
	builder := c.store.NewEntryQueryBuilder(user.ID)
	builder.WithFeedID(feed.ID)
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(user.EntryDirection)
	builder.WithOffset(offset)
	builder.WithLimit(nbItemsPerPage)

	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, err)
		return
	}

	sess := session.New(c.store, request.SessionID(r))
	view := view.New(c.tpl, r, sess)
	view.Set("feed", feed)
	view.Set("entries", entries)
	view.Set("total", count)
	view.Set("pagination", c.getPagination(route.Path(c.router, "feedEntries", "feedID", feed.ID), count, offset))
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", c.store.CountErrorFeeds(user.ID))
	view.Set("hasSaveEntry", c.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("feed_entries"))
}
