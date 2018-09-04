// BSD 2-Clause License
//
// Copyright (c) 2018, Krzysztof Kowalczyk
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package notiontypes

import (
	"encoding/json"
	"time"
)

// BlockWithRole describes a block info
type BlockWithRole struct {
	Role  string `json:"role"`
	Value *Block `json:"value"`
}

// Block describes a block
type Block struct {
	// values that come from JSON
	// if false, the page is deleted
	Alive bool `json:"alive"`
	// List of block ids for that make up content of this block
	// Use Content to get corresponding block (they are in the same order)
	ContentIDs   []string `json:"content,omitempty"`
	CopiedFrom   string   `json:"copied_from,omitempty"`
	CollectionID string   `json:"collection_id,omitempty"` // for BlockCollectionView
	// ID of the user who created this block
	CreatedBy   string `json:"created_by"`
	CreatedTime int64  `json:"created_time"`
	// List of block ids with discussion content
	DiscussionIDs []string `json:"discussion,omitempty"`
	// those ids seem to map to storage in s3
	// https://s3-us-west-2.amazonaws.com/secure.notion-static.com/${id}/${name}
	FileIDs   []string        `json:"file_ids,omitempty"`
	FormatRaw json.RawMessage `json:"format,omitempty"`
	// a unique ID of the block
	ID string `json:"id"`

	// TODO: don't know what this means
	IgnoreBlockCount bool `json:"ignore_block_count,omitempty"`

	// ID of the user who last edited this block
	LastEditedBy   string `json:"last_edited_by"`
	LastEditedTime int64  `json:"last_edited_time"`
	// ID of parent Block
	ParentID    string `json:"parent_id"`
	ParentTable string `json:"parent_table"`
	// not always available
	Permissions *[]Permission          `json:"permissions,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	// type of the block e.g. TypeText, TypePage etc.
	Type string `json:"type"`
	// blocks are versioned
	Version int64    `json:"version"`
	ViewIDs []string `json:"view_ids,omitempty"`

	// Values calculated by us

	// maps ContentIDs array
	Content []*Block `json:"content_resolved,omitempty"`
	// this is for some types like TypePage, TypeText, TypeHeader etc.
	InlineContent []*InlineBlock `json:"inline_content,omitempty"`

	// for BlockPage
	Title string `json:"title,omitempty"`

	// For BlockTodo, a checked state
	IsChecked bool `json:"is_checked,omitempty"`

	// for BlockBookmark
	Description string `json:"description,omitempty"`
	Link        string `json:"link,omitempty"`

	// for BlockBookmark it's the url of the page
	// for BlockGist it's the url for the gist
	// fot BlockImage it's url of the image, but use ImageURL instead
	// because Source is sometimes not accessible
	// for BlockFile it's url of the file
	Source string `json:"source,omitempty"`

	// for BlockFile
	FileSize string `json:"file_size,omitempty"`

	// for BlockImage it's an URL built from Source that is always accessible
	ImageURL string `json:"image_url,omitempty"`

	// for BlockCode
	Code         string `json:"code,omitempty"`
	CodeLanguage string `json:"code_language,omitempty"`

	// for BlockCollectionView
	// It looks like the info about which view is selected is stored in browser
	CollectionViews []*CollectionViewInfo `json:"collection_views,omitempty"`

	FormatPage     *FormatPage     `json:"format_page,omitempty"`
	FormatBookmark *FormatBookmark `json:"format_bookmark,omitempty"`
	FormatImage    *FormatImage    `json:"format_image,omitempty"`
	FormatColumn   *FormatColumn   `json:"format_column,omitempty"`
	FormatText     *FormatText     `json:"format_text,omitempty"`
	FormatTable    *FormatTable    `json:"format_table,omitempty"`
	FormatVideo    *FormatVideo    `json:"format_video,omitempty"`
}

// CollectionViewInfo describes a particular view of the collection
type CollectionViewInfo struct {
	CollectionView *CollectionView
	Collection     *Collection
	CollectionRows []*Block
}

// CreatedOn return the time the page was created
func (b *Block) CreatedOn() time.Time {
	return time.Unix(b.CreatedTime/1000, 0)
}

// UpdatedOn returns the time the page was last updated
func (b *Block) UpdatedOn() time.Time {
	return time.Unix(b.LastEditedTime/1000, 0)
}

// IsLinkToPage returns true if block element is a link to existing page
// (as opposed to )
func (b *Block) IsLinkToPage() bool {
	if b.Type != BlockPage {
		return false
	}
	return b.ParentTable == TableSpace
}

// IsPage returns true if block represents a page (either a sub-page or
// a link to a page)
func (b *Block) IsPage() bool {
	return b.Type == BlockPage
}

// IsImage returns true if block represents an image
func (b *Block) IsImage() bool {
	return b.Type == BlockImage
}

// IsCode returns true if block represents a code block
func (b *Block) IsCode() bool {
	return b.Type == BlockCode
}

// FormatPage describes format for TypePage
type FormatPage struct {
	// /images/page-cover/gradients_11.jpg
	PageCover string `json:"page_cover"`
	// e.g. 0.6
	PageCoverPosition float64 `json:"page_cover_position"`
	PageFont          string  `json:"page_font"`
	PageFullWidth     bool    `json:"page_full_width"`
	// it's url like https://s3-us-west-2.amazonaws.com/secure.notion-static.com/8b3930e3-9dfe-4ba7-a845-a8ff69154f2a/favicon-256.png
	// or emoji like "✉️"
	PageIcon      string `json:"page_icon"`
	PageSmallText bool   `json:"page_small_text"`

	// calculated by us
	PageCoverURL string `json:"page_cover_url,omitempty"`
}

// FormatBookmark describes format for BlockBookmark
type FormatBookmark struct {
	BookmarkIcon string `json:"bookmark_icon"`
}

// FormatImage describes format for BlockImage
type FormatImage struct {
	// comes from notion API
	BlockAspectRatio   float64 `json:"block_aspect_ratio"`
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
	BlockWidth         float64 `json:"block_width"`
	DisplaySource      string  `json:"display_source,omitempty"`

	// calculated by us
	ImageURL string `json:"image_url,omitempty"`
}

// FormatVideo describes fromat form TypeVideo
type FormatVideo struct {
	BlockWidth         int64   `json:"block_width"`
	BlockHeight        int64   `json:"block_height"`
	DisplaySource      string  `json:"display_source"`
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockAspectRatio   float64 `json:"block_aspect_ratio"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
}

// FormatText describes format for TypeText
// TODO: possibly more?
type FormatText struct {
	BlockColor *string `json:"block_color,omitempty"`
}

// FormatTable describes format for TypeTable
type FormatTable struct {
	TableWrap       bool             `json:"table_wrap"`
	TableProperties []*TableProperty `json:"table_properties"`
}

// TableProperty describes property of a table
type TableProperty struct {
	Width    int    `json:"width"`
	Visible  bool   `json:"visible"`
	Property string `json:"property"`
}

// FormatColumn describes format for TypeColumn
type FormatColumn struct {
	ColumnRation float64 `json:"column_ratio"` // e.g. 0.5 for half-sized column
}

// Permission describes user permissions
type Permission struct {
	Role   string  `json:"role"`
	Type   string  `json:"type"`
	UserID *string `json:"user_id,omitempty"`
}
