package gosteno

type nullSink struct {
	records []*Record
}

func newNullSink() *nullSink {
	nSink := new(nullSink)
	nSink.records = make([]*Record, 0, 10)
	return nSink
}

func (nSink *nullSink) AddRecord(record *Record) {
	nSink.records = append(nSink.records, record)
}

func (nSink *nullSink) Flush() {

}

func (nSink *nullSink) SetCodec(codec Codec) {

}

func (nSink *nullSink) GetCodec() Codec {
	return nil
}
