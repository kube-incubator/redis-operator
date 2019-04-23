package syncer

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// SyncResult is a result of an Sync call
type SyncResult struct {
	Operation    controllerutil.OperationResult
	EventType    string
	EventReason  string
	EventMessage string
}

// SetEventData sets event data on an SyncResult
func (r *SyncResult) SetEventData(eventType, reason, message string) {
	r.EventType = eventType
	r.EventReason = reason
	r.EventMessage = message
}

// Interface represents a syncer. A syncer persists an object
// (known as subject), into a store (kubernetes apiserver or generic stores)
// and records kubernetes events
type Interface interface {
	// GetObject returns the object for which sync applies
	GetObject() interface{}
	// GetOwner returns the object owner or nil if object does not have one
	GetOwner() runtime.Object
	// Sync persists data into the external store
	Sync(context.Context) (SyncResult, error)
}
