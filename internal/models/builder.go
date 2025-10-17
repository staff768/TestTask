package models

import "fmt"

type QueryBuilderInterface interface {
	WithStartDate(startDate *string) *QueryBuilder
	WithEndDate(endDate *string) *QueryBuilder
	WithUserId(userId *string) *QueryBuilder
	WithServiceName(serviceName *string) *QueryBuilder
	WithName(name *string) *QueryBuilder
	BuildQuery() string
}

type QueryBuilder struct {
	Query       string
	placeHolder int
	Args        []interface{}
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		Query: "SELECT COALESCE(SUM(price), 0) FROM subsciptions WHERE 1=1",
		Args:  []interface{}{},
	}
}

func (builder *QueryBuilder) WithStartDate(startDate *string) *QueryBuilder {
	if startDate != nil && *startDate != "" {
		builder.placeHolder++
		builder.Query = builder.Query + fmt.Sprintf(" AND start_date >= $%d", builder.placeHolder)
		builder.Args = append(builder.Args, *startDate)
	}
	return builder
}
func (builder *QueryBuilder) WithEndDate(endDate *string) *QueryBuilder {
	if endDate != nil && *endDate != "" {
		builder.placeHolder++
		builder.Query = builder.Query + fmt.Sprintf(" AND end_date <= $%d", builder.placeHolder)
		builder.Args = append(builder.Args, *endDate)
	}
	return builder
}

func (builder *QueryBuilder) WithUserId(userId *string) *QueryBuilder {
	if userId != nil && *userId != "" {
		builder.placeHolder++
		builder.Query = builder.Query + fmt.Sprintf(" AND user_id = $%d", builder.placeHolder)
		builder.Args = append(builder.Args, *userId)
	}
	return builder
}

func (builder *QueryBuilder) WithServiceName(serviceName *string) *QueryBuilder {
	if serviceName != nil && *serviceName != "" {
		builder.placeHolder++
		builder.Query = builder.Query + fmt.Sprintf(" AND service_name = $%d", builder.placeHolder)
		builder.Args = append(builder.Args, *serviceName)
	}
	return builder
}

func (builder *QueryBuilder) BuildQuery() (string, []interface{}) {
	return builder.Query, builder.Args
}
