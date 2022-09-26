package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// This is a wrapper core that makes sure that pre-specified fields are unique
type uniqueFieldsCore struct {
	root                zapcore.Core
	current             zapcore.Core
	fields              []zapcore.Field
	reorderLoggedFields bool
}

func MakeFieldsUnique(reorderLoggedFields bool) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &uniqueFieldsCore{
			root:                core,
			current:             core,
			reorderLoggedFields: reorderLoggedFields,
		}
	})
}

func (u uniqueFieldsCore) Enabled(level zapcore.Level) bool {
	return u.current.Enabled(level)
}

func (u uniqueFieldsCore) With(newFields []zapcore.Field) zapcore.Core {
	newFieldList := u.makeUniqueFields(u.fields, newFields)
	newCore := u.root
	if !u.reorderLoggedFields {
		newCore = u.root.With(newFieldList)
	}

	return &uniqueFieldsCore{
		root:                u.root,
		current:             newCore,
		fields:              newFieldList,
		reorderLoggedFields: u.reorderLoggedFields,
	}
}

func (u uniqueFieldsCore) makeUniqueFields(curFields []zapcore.Field,
	newFields []zapcore.Field) []zapcore.Field {

	newFieldList := make([]zapcore.Field, 0, len(curFields)+len(newFields))
	newFieldList = append(newFieldList, newFields...)

outer:
	for _, f := range curFields {
		// Skip all the existing fields with the names that match one
		// of the new fields.
		for _, k := range newFields {
			if f.Key == k.Key {
				continue outer
			}
		}
		newFieldList = append(newFieldList, f)
	}
	return newFieldList
}

func (u uniqueFieldsCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	enabler, ok := u.current.(zapcore.LevelEnabler)
	if ok {
		if enabler.Enabled(entry.Level) {
			return checked.AddCore(entry, u)
		}
		return nil
	} else {
		return checked.AddCore(entry, u)
	}
}

func (u uniqueFieldsCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if u.reorderLoggedFields {
		reordered := u.makeUniqueFields(u.fields, fields)
		return u.current.Write(entry, reordered)
	} else {
		return u.current.Write(entry, fields)
	}
}

func (u uniqueFieldsCore) Sync() error {
	return u.current.Sync()
}
