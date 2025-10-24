package sqlbuilder

// BuilderFactory 构建器工厂
type BuilderFactory struct {
	dbType DatabaseType
}

// NewBuilderFactory 创建构建器工厂
func NewBuilderFactory(dbType DatabaseType) *BuilderFactory {
	return &BuilderFactory{
		dbType: dbType,
	}
}

// CreateQueryBuilder 创建查询构建器
func (f *BuilderFactory) CreateQueryBuilder() QueryBuilder {
	switch f.dbType {
	case Dameng:
		return NewUniversalBuilder(Dameng)
	case MySQL:
		return NewUniversalBuilder(MySQL)
	case SQLite:
		return NewUniversalBuilder(SQLite)
	case Postgres:
		return NewUniversalBuilder(Postgres)
	default:
		return NewUniversalBuilder(MySQL) // 默认使用MySQL
	}
}

// CreateDamengBuilder 创建达梦专用构建器
func (f *BuilderFactory) CreateDamengBuilder() *DamengQueryBuilder {
	return NewDamengBuilder()
}

// GetDatabaseType 获取数据库类型
func (f *BuilderFactory) GetDatabaseType() DatabaseType {
	return f.dbType
}
