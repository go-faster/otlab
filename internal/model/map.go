package model

import (
	"github.com/go-faster/jx"
	"go.opentelemetry.io/otel/attribute"
)

// MarshalAttrSet marshals an attribute.Set to JSON.
func MarshalAttrSet(set attribute.Set) ([]byte, error) {
	e := &jx.Encoder{}
	err := EncodeAttrSet(set, e)
	if err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

// EncodeAttrSet encodes an attribute.Set to JSON.
func EncodeAttrSet(set attribute.Set, e *jx.Encoder) error {
	var retErr error
	e.Obj(func(e *jx.Encoder) {
		d := jx.Decoder{}
		iter := set.Iter()
		for iter.Next() {
			kv := iter.Attribute()
			if retErr != nil {
				break
			}
			e.Field(string(kv.Key), func(e *jx.Encoder) {
				data, err := kv.Value.MarshalJSON()
				if err != nil {
					retErr = err
					return
				}
				d.ResetBytes(data)
				retErr = d.Obj(func(d *jx.Decoder, key string) error {
					switch key {
					default:
						return d.Skip()
					case "Value":
						raw, err := d.Raw()
						if err != nil {
							return err
						}
						e.Raw(raw)
						return nil
					}
				})
			})
		}
	})
	return retErr
}
