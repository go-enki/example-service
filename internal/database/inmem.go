package database

type InMem struct {
	
}

func NewInMemDB() *InMem {
	return &InMem{}
}

func (db InMem) GetHelloTemplate(name string) (template string, err error) {
	return "Ohai there, %s", nil
}