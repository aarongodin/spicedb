// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: authzed/api/v0/core.proto

package v0

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
)

// Validate checks the field values on RelationTuple with the rules defined in
// the proto definition for this message. If any rules are violated, an error
// is returned.
func (m *RelationTuple) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetObjectAndRelation() == nil {
		return RelationTupleValidationError{
			field:  "ObjectAndRelation",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetObjectAndRelation()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return RelationTupleValidationError{
				field:  "ObjectAndRelation",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetUser() == nil {
		return RelationTupleValidationError{
			field:  "User",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetUser()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return RelationTupleValidationError{
				field:  "User",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// RelationTupleValidationError is the validation error returned by
// RelationTuple.Validate if the designated constraints aren't met.
type RelationTupleValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RelationTupleValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RelationTupleValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RelationTupleValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RelationTupleValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RelationTupleValidationError) ErrorName() string { return "RelationTupleValidationError" }

// Error satisfies the builtin error interface
func (e RelationTupleValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRelationTuple.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RelationTupleValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RelationTupleValidationError{}

// Validate checks the field values on ObjectAndRelation with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *ObjectAndRelation) Validate() error {
	if m == nil {
		return nil
	}

	if len(m.GetNamespace()) > 128 {
		return ObjectAndRelationValidationError{
			field:  "Namespace",
			reason: "value length must be at most 128 bytes",
		}
	}

	if !_ObjectAndRelation_Namespace_Pattern.MatchString(m.GetNamespace()) {
		return ObjectAndRelationValidationError{
			field:  "Namespace",
			reason: "value does not match regex pattern \"^([a-z][a-z0-9_]{2,62}[a-z0-9]/)?[a-z][a-z0-9_]{2,62}[a-z0-9]$\"",
		}
	}

	if len(m.GetObjectId()) > 64 {
		return ObjectAndRelationValidationError{
			field:  "ObjectId",
			reason: "value length must be at most 64 bytes",
		}
	}

	if !_ObjectAndRelation_ObjectId_Pattern.MatchString(m.GetObjectId()) {
		return ObjectAndRelationValidationError{
			field:  "ObjectId",
			reason: "value does not match regex pattern \"^[a-zA-Z0-9/_-]{2,64}$\"",
		}
	}

	if len(m.GetRelation()) > 64 {
		return ObjectAndRelationValidationError{
			field:  "Relation",
			reason: "value length must be at most 64 bytes",
		}
	}

	if !_ObjectAndRelation_Relation_Pattern.MatchString(m.GetRelation()) {
		return ObjectAndRelationValidationError{
			field:  "Relation",
			reason: "value does not match regex pattern \"^(\\\\.\\\\.\\\\.|[a-z][a-z0-9_]{2,62}[a-z0-9])$\"",
		}
	}

	return nil
}

// ObjectAndRelationValidationError is the validation error returned by
// ObjectAndRelation.Validate if the designated constraints aren't met.
type ObjectAndRelationValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ObjectAndRelationValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ObjectAndRelationValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ObjectAndRelationValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ObjectAndRelationValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ObjectAndRelationValidationError) ErrorName() string {
	return "ObjectAndRelationValidationError"
}

// Error satisfies the builtin error interface
func (e ObjectAndRelationValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sObjectAndRelation.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ObjectAndRelationValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ObjectAndRelationValidationError{}

var _ObjectAndRelation_Namespace_Pattern = regexp.MustCompile("^([a-z][a-z0-9_]{2,62}[a-z0-9]/)?[a-z][a-z0-9_]{2,62}[a-z0-9]$")

var _ObjectAndRelation_ObjectId_Pattern = regexp.MustCompile("^[a-zA-Z0-9/_-]{2,64}$")

var _ObjectAndRelation_Relation_Pattern = regexp.MustCompile("^(\\.\\.\\.|[a-z][a-z0-9_]{2,62}[a-z0-9])$")

// Validate checks the field values on RelationReference with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *RelationReference) Validate() error {
	if m == nil {
		return nil
	}

	if len(m.GetNamespace()) > 128 {
		return RelationReferenceValidationError{
			field:  "Namespace",
			reason: "value length must be at most 128 bytes",
		}
	}

	if !_RelationReference_Namespace_Pattern.MatchString(m.GetNamespace()) {
		return RelationReferenceValidationError{
			field:  "Namespace",
			reason: "value does not match regex pattern \"^([a-z][a-z0-9_]{2,62}[a-z0-9]/)?[a-z][a-z0-9_]{2,62}[a-z0-9]$\"",
		}
	}

	if len(m.GetRelation()) > 64 {
		return RelationReferenceValidationError{
			field:  "Relation",
			reason: "value length must be at most 64 bytes",
		}
	}

	if !_RelationReference_Relation_Pattern.MatchString(m.GetRelation()) {
		return RelationReferenceValidationError{
			field:  "Relation",
			reason: "value does not match regex pattern \"^(\\\\.\\\\.\\\\.|[a-z][a-z0-9_]{2,62}[a-z0-9])$\"",
		}
	}

	return nil
}

// RelationReferenceValidationError is the validation error returned by
// RelationReference.Validate if the designated constraints aren't met.
type RelationReferenceValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RelationReferenceValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RelationReferenceValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RelationReferenceValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RelationReferenceValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RelationReferenceValidationError) ErrorName() string {
	return "RelationReferenceValidationError"
}

// Error satisfies the builtin error interface
func (e RelationReferenceValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRelationReference.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RelationReferenceValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RelationReferenceValidationError{}

var _RelationReference_Namespace_Pattern = regexp.MustCompile("^([a-z][a-z0-9_]{2,62}[a-z0-9]/)?[a-z][a-z0-9_]{2,62}[a-z0-9]$")

var _RelationReference_Relation_Pattern = regexp.MustCompile("^(\\.\\.\\.|[a-z][a-z0-9_]{2,62}[a-z0-9])$")

// Validate checks the field values on User with the rules defined in the proto
// definition for this message. If any rules are violated, an error is returned.
func (m *User) Validate() error {
	if m == nil {
		return nil
	}

	switch m.UserOneof.(type) {

	case *User_Userset:

		if m.GetUserset() == nil {
			return UserValidationError{
				field:  "Userset",
				reason: "value is required",
			}
		}

		if v, ok := interface{}(m.GetUserset()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return UserValidationError{
					field:  "Userset",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	default:
		return UserValidationError{
			field:  "UserOneof",
			reason: "value is required",
		}

	}

	return nil
}

// UserValidationError is the validation error returned by User.Validate if the
// designated constraints aren't met.
type UserValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e UserValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e UserValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e UserValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e UserValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e UserValidationError) ErrorName() string { return "UserValidationError" }

// Error satisfies the builtin error interface
func (e UserValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sUser.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = UserValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = UserValidationError{}

// Validate checks the field values on Zookie with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Zookie) Validate() error {
	if m == nil {
		return nil
	}

	if len(m.GetToken()) < 1 {
		return ZookieValidationError{
			field:  "Token",
			reason: "value length must be at least 1 bytes",
		}
	}

	return nil
}

// ZookieValidationError is the validation error returned by Zookie.Validate if
// the designated constraints aren't met.
type ZookieValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ZookieValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ZookieValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ZookieValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ZookieValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ZookieValidationError) ErrorName() string { return "ZookieValidationError" }

// Error satisfies the builtin error interface
func (e ZookieValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sZookie.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ZookieValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ZookieValidationError{}

// Validate checks the field values on RelationTupleUpdate with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *RelationTupleUpdate) Validate() error {
	if m == nil {
		return nil
	}

	if _, ok := RelationTupleUpdate_Operation_name[int32(m.GetOperation())]; !ok {
		return RelationTupleUpdateValidationError{
			field:  "Operation",
			reason: "value must be one of the defined enum values",
		}
	}

	if m.GetTuple() == nil {
		return RelationTupleUpdateValidationError{
			field:  "Tuple",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetTuple()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return RelationTupleUpdateValidationError{
				field:  "Tuple",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// RelationTupleUpdateValidationError is the validation error returned by
// RelationTupleUpdate.Validate if the designated constraints aren't met.
type RelationTupleUpdateValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RelationTupleUpdateValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RelationTupleUpdateValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RelationTupleUpdateValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RelationTupleUpdateValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RelationTupleUpdateValidationError) ErrorName() string {
	return "RelationTupleUpdateValidationError"
}

// Error satisfies the builtin error interface
func (e RelationTupleUpdateValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRelationTupleUpdate.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RelationTupleUpdateValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RelationTupleUpdateValidationError{}

// Validate checks the field values on RelationTupleTreeNode with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *RelationTupleTreeNode) Validate() error {
	if m == nil {
		return nil
	}

	if v, ok := interface{}(m.GetExpanded()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return RelationTupleTreeNodeValidationError{
				field:  "Expanded",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	switch m.NodeType.(type) {

	case *RelationTupleTreeNode_IntermediateNode:

		if v, ok := interface{}(m.GetIntermediateNode()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return RelationTupleTreeNodeValidationError{
					field:  "IntermediateNode",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *RelationTupleTreeNode_LeafNode:

		if v, ok := interface{}(m.GetLeafNode()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return RelationTupleTreeNodeValidationError{
					field:  "LeafNode",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// RelationTupleTreeNodeValidationError is the validation error returned by
// RelationTupleTreeNode.Validate if the designated constraints aren't met.
type RelationTupleTreeNodeValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RelationTupleTreeNodeValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RelationTupleTreeNodeValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RelationTupleTreeNodeValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RelationTupleTreeNodeValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RelationTupleTreeNodeValidationError) ErrorName() string {
	return "RelationTupleTreeNodeValidationError"
}

// Error satisfies the builtin error interface
func (e RelationTupleTreeNodeValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRelationTupleTreeNode.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RelationTupleTreeNodeValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RelationTupleTreeNodeValidationError{}

// Validate checks the field values on SetOperationUserset with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *SetOperationUserset) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Operation

	for idx, item := range m.GetChildNodes() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return SetOperationUsersetValidationError{
					field:  fmt.Sprintf("ChildNodes[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// SetOperationUsersetValidationError is the validation error returned by
// SetOperationUserset.Validate if the designated constraints aren't met.
type SetOperationUsersetValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SetOperationUsersetValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SetOperationUsersetValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SetOperationUsersetValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SetOperationUsersetValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SetOperationUsersetValidationError) ErrorName() string {
	return "SetOperationUsersetValidationError"
}

// Error satisfies the builtin error interface
func (e SetOperationUsersetValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSetOperationUserset.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SetOperationUsersetValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SetOperationUsersetValidationError{}

// Validate checks the field values on DirectUserset with the rules defined in
// the proto definition for this message. If any rules are violated, an error
// is returned.
func (m *DirectUserset) Validate() error {
	if m == nil {
		return nil
	}

	for idx, item := range m.GetUsers() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return DirectUsersetValidationError{
					field:  fmt.Sprintf("Users[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// DirectUsersetValidationError is the validation error returned by
// DirectUserset.Validate if the designated constraints aren't met.
type DirectUsersetValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DirectUsersetValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DirectUsersetValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DirectUsersetValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DirectUsersetValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DirectUsersetValidationError) ErrorName() string { return "DirectUsersetValidationError" }

// Error satisfies the builtin error interface
func (e DirectUsersetValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDirectUserset.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DirectUsersetValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DirectUsersetValidationError{}