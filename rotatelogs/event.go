package rotatelogs

// Handle 处理事件
func (h HandlerFunc) Handle(e Event) {
	h(e)
}

// Type 获取事件类型
func (e *FileRotatedEvent) Type() EventType {
	return FileRotatedEventType
}

// PreviousFile 获取轮转前的文件名
func (e *FileRotatedEvent) PreviousFile() string {
	return e.prev
}

// CurrentFile 获取轮转后的当前文件名
func (e *FileRotatedEvent) CurrentFile() string {
	return e.current
}
