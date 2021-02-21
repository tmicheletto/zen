package file

import "io/ioutil"

type Service struct{}

func New() *Service {
	return &Service{}
}

func (svc *Service) ReadFile(fileName string) ([]byte, error) {
	return ioutil.ReadFile(fileName)
}
