package mwsettings

/*
SettingDecoder is the interface to which all setting decoders comply. A setting decoder
is a struct with methods for decoding JSON setting information.
*/
type SettingDecoder interface {
	DecodeSetting(s interface{}) (string, interface{})
	CanDecodeSetting(path string) bool
}

/*
BasicSettingDecoder is a dumb decoder that simpily decodes things by returning them. It will only
decode if path matches.
*/
type BasicSettingDecoder struct {
	path string
}

/*
DecodeSetting decodes a setting by simpily returning it.
return:
  setting name as dictated by path name,
  the interface that was passed in.
*/
func (dec *BasicSettingDecoder) DecodeSetting(s interface{}) (string, interface{}) {
	return dec.path, s
}

/*
CanDecodeSetting returns true if the objects internal path param matches path
*/
func (dec *BasicSettingDecoder) CanDecodeSetting(path string) bool {
	if dec.path == path {
		return true
	}
	return false
}

/*
NewBasicDecoder returns a decoder that performs dumb decoding on any setting matching path.
An example path would be: "general/TCPProtocol".
*/
func NewBasicDecoder(path string) SettingDecoder {
	return &BasicSettingDecoder{path}
}

/*
FunctionalSettingDecoder a setting decoder that takes user provided functions for both
decode and canDecode operations.
*/
type FunctionalSettingDecoder struct {
	canDecodeFunc func(path string) bool
	decodeFunc    func(s interface{}) (string, interface{})
}

/*
DecodeSetting decode the setting item based on user provided function
*/
func (dec *FunctionalSettingDecoder) DecodeSetting(s interface{}) (string, interface{}) {
	return dec.decodeFunc(s)
}

/*
CanDecodeSetting decide if this object can handle this setting path based on user provided function
*/
func (dec *FunctionalSettingDecoder) CanDecodeSetting(path string) bool {
	return dec.canDecodeFunc(path)
}

/*
NewFunctionalSettingDecoder constructs a new functional decoder with the provided user decode and canDecode functions
*/
func NewFunctionalSettingDecoder(decodeFunc func(s interface{}) (string, interface{}),
	canDecodeFunc func(path string) bool) SettingDecoder {
	return &FunctionalSettingDecoder{canDecodeFunc, decodeFunc}
}
