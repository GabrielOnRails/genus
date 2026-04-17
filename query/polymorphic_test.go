package query

import (
	"reflect"
	"testing"

	"github.com/go-genus/genus/core"
)

// --- Polymorphic test models ---

type testComment struct {
	core.Model
	Body            string `db:"body"`
	CommentableType string `db:"commentable_type"`
	CommentableID   int64  `db:"commentable_id"`
}

func (testComment) TableName() string { return "comments" }

type testArticle struct {
	core.Model
	Title    string        `db:"title"`
	Comments []testComment `relation:"polymorphic,polymorphic=commentable"`
}

func (testArticle) TableName() string { return "articles" }

// --- Tests ---

func TestIsPolymorphicHasMany_ParentSide(t *testing.T) {
	// testArticle does NOT have commentable_type/commentable_id fields
	// so it should be detected as the has_many side
	meta := &core.RelationshipMeta{
		Type:            core.Polymorphic,
		FieldName:       "Comments",
		Polymorphic:     "commentable",
		PolymorphicType: "commentable_type",
		PolymorphicID:   "commentable_id",
	}

	result := isPolymorphicHasMany[testArticle](meta)
	if !result {
		t.Error("expected testArticle to be detected as has_many side of polymorphic relationship")
	}
}

func TestIsPolymorphicHasMany_ChildSide(t *testing.T) {
	// testComment HAS commentable_type and commentable_id fields
	// so it should NOT be detected as has_many
	meta := &core.RelationshipMeta{
		Type:            core.Polymorphic,
		FieldName:       "Commentable",
		Polymorphic:     "commentable",
		PolymorphicType: "commentable_type",
		PolymorphicID:   "commentable_id",
	}

	result := isPolymorphicHasMany[testComment](meta)
	if result {
		t.Error("expected testComment to NOT be detected as has_many side")
	}
}

func TestToSnakeCasePlural_Polymorphic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Post", "post"},
		{"BlogPost", "blog_post"},
		{"User", "user"},
		{"CommentThread", "comment_thread"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCasePlural(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCasePlural(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase_Polymorphic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_id", "UserId"},
		{"commentable_type", "CommentableType"},
		{"commentable_id", "CommentableId"},
		{"id", "Id"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetTableNameFromType_PolymorphicStruct(t *testing.T) {
	typ := reflect.TypeOf(testArticle{})
	name := getTableNameFromType(typ)
	if name != "test_article" {
		t.Errorf("expected 'test_article', got %q", name)
	}
}

func TestGetTableNameFromType_PolymorphicSlice(t *testing.T) {
	typ := reflect.TypeOf([]testComment{})
	name := getTableNameFromType(typ)
	if name != "test_comment" {
		t.Errorf("expected 'test_comment', got %q", name)
	}
}

func TestParseRelationTag_Polymorphic(t *testing.T) {
	// Simulate parsing polymorphic tag
	field := reflect.TypeOf(testArticle{}).Field(2) // Comments field
	tag := field.Tag.Get("relation")

	if tag == "" {
		t.Fatal("expected relation tag on Comments field")
	}

	meta, err := core.ParseRelationTag(field, tag)
	if err != nil {
		t.Fatalf("ParseRelationTag failed: %v", err)
	}

	if meta.Type != core.Polymorphic {
		t.Errorf("expected Polymorphic type, got %s", meta.Type)
	}
	if meta.Polymorphic != "commentable" {
		t.Errorf("expected polymorphic=commentable, got %s", meta.Polymorphic)
	}
	if meta.PolymorphicType != "commentable_type" {
		t.Errorf("expected polymorphic_type=commentable_type, got %s", meta.PolymorphicType)
	}
	if meta.PolymorphicID != "commentable_id" {
		t.Errorf("expected polymorphic_id=commentable_id, got %s", meta.PolymorphicID)
	}
}
