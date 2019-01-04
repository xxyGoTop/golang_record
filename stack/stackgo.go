package stackgo

//声明一个栈
type Stack struct {
	size             int
	currentPage      []interface{}
	pages            [][]interface{}
	offset           int
	capacity         int
	pageSize         int
	currentPageIndex int
}

const s_DefaultAllocPageSize = 4096

//入栈
func (s *Stack) Push(elem ...interface{}) {
	if elem == nil || len(elem) == 0 {
		return
	}

	//根据新增加的data, 更新stack的大小
	if s.size+len(elem) > s.capacity {
		newPages := len(elem) / s.pageSize
		if len(elem)%s.pageSize != 0 {
			newPages++
		}

		for newPages > 0 {
			page := make([]interface{}, s.pageSize)
			s.pages = append(s.pages, page)
			s.capacity += len(page)
			newPages--
		}
	}

	//向stack里增加data
	//根据当前页的大小来 判断是否插入数据
	//增加扩容
	s.size += len(elem)
	for len(elem) > 0 {
		available := len(s.currentPage) - s.offset
		min := len(elem)
		if available < min {
			min = available
		}
		copy(s.currentPage[s.offset:], elem[:min])
		elem = elem[min:]
		s.offset += min
		if len(elem) > 0 {
			s.currentPage = s.pages[s.currentPageIndex+1]
			s.currentPageIndex++
			s.offset = 0
		}
	}
}

//出栈
func (s *Stack) Pop() interface{} {
	if s.size <= 0 {
		return nil
	}

	s.offset--
	s.size--

	if s.offset < 0 {
		s.offset = s.pageSize - 1
		s.currentPage, s.pages = s.pages[s.currentPageIndex-1], s.pages[:s.currentPageIndex+1]
		s.capacity -= s.pageSize
		s.currentPageIndex--
	}

	return s.currentPage[s.offset]
}

//栈大小
func (s *Stack) Size() int {
	return s.size
}

//创建一个stack
func NewStack(cap int) *Stack {
	var defaultPageSize int
	var stack = new(Stack)

	if cap == 0 {
		defaultPageSize = s_DefaultAllocPageSize
	} else {
		defaultPageSize = cap
	}
	stack.currentPage = make([]interface{}, defaultPageSize)
	stack.pages = [][]interface{}{stack.currentPage}
	stack.offset = 0
	stack.capacity = defaultPageSize
	stack.pageSize = defaultPageSize
	stack.size = 0
	stack.currentPageIndex = 0

	return stack
}
