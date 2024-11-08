package db

type BeforeCreate interface {
	BeforeCreate(data Keyable) error
}

type AfterCreate interface {
	AfterCreate(data Keyable) error
}

type BeforeDelete interface {
	BeforeDelete(data Keyable) error
}

type AfterDelete interface {
	AfterDelete(data Keyable) error
}

type BeforeUpdate interface {
	BeforeUpdate(data Keyable) error
}

type AfterUpdate interface {
	AfterUpdate(data Keyable) error
}
