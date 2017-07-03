package ansiblelocal

import (
	"github.com/hashicorp/packer/packer"
	"io"
	"os"
)

type expectation struct {
	description string
	call        interface{}
	satisfied   bool
}

func (e *expectation) markSatisfied() {
	e.satisfied = true
}

type mock struct {
	expects []expectation
	current int
	strict  bool
}

func (m *mock) currentExpect() *expectation {
	if m.expects[m.current].satisfied {
		m.current++
	}
	if m.current == len(m.expects) {
		panic("All expectations already satisfied")
	}
	return &m.expects[m.current]
}

func (m *mock) appendExpectation(description string, call interface{}) {
	m.expects = append(m.expects, expectation{description: description, call: call})
}

func (m *mock) isSatisfied() bool {
	return m.current == (len(m.expects)-1) && m.expects[m.current].satisfied
}

func (m *mock) verify() {
	for _, expect := range m.expects {
		if !expect.satisfied {
			panic(expect.description)
		}
	}
}

type startMock struct {
	mock
}
type startExpect func(*expectation, *packer.RemoteCmd) error

func (e *startMock) call(cmd *packer.RemoteCmd) error {
	if !e.strict && e.isSatisfied() {
		return nil
	}
	return e.currentExpect().call.(startExpect)(e.currentExpect(), cmd)
}

type uploadMock struct {
	mock
}
type uploadExpect func(*expectation, string, io.Reader, *os.FileInfo) error

func (e *uploadMock) call(dst string, contents io.Reader, info *os.FileInfo) error {
	if !e.strict && e.isSatisfied() {
		return nil
	}
	return e.currentExpect().call.(uploadExpect)(e.currentExpect(), dst, contents, info)
}

type communicatorMock struct {
	startMock         startMock
	uploadMock        uploadMock
}


func (c *communicatorMock) expectStart(description string, expect startExpect) {
	c.startMock.appendExpectation(description, expect)
}

func (c *communicatorMock) expectUpload(description string, expect uploadExpect) {
	c.uploadMock.appendExpectation(description, expect)
}

func (c *communicatorMock) Start(cmd *packer.RemoteCmd) error {
	err := c.startMock.call(cmd)
	if !cmd.Exited {
		cmd.SetExited(0)
	}
	return err
}

func (c *communicatorMock) Upload(dst string, contents io.Reader, info *os.FileInfo) error {
	return c.uploadMock.call(dst, contents, info)
}

func (c *communicatorMock) UploadDir(dst, src string, exclude []string) error {
	return nil
}

func (c *communicatorMock) Download(src string, dst io.Writer) error {
	return nil
}

func (c *communicatorMock) DownloadDir(src, dst string, exclude []string) error {
	return nil
}

func (c *communicatorMock) verify() {
	c.startMock.verify()
	c.uploadMock.verify()
}
