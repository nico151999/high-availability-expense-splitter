package environment

import (
	"context"
	"fmt"
)

// GetGroupServerPort returns the port the group service will run on
func GetGroupServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "GROUP_SERVER_PORT")
}

// GetPersonServerPort returns the port the person service will run on
func GetPersonServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "PERSON_SERVER_PORT")
}

// GetCategoryServerPort returns the port the category service will run on
func GetCategoryServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "CATEGORY_SERVER_PORT")
}

// GetExpenseServerPort returns the port the expense service will run on
func GetExpenseServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "EXPENSE_SERVER_PORT")
}

// GetExpensestakeServerPort returns the port the expense service will run on
func GetExpensestakeServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "EXPENSESTAKE_SERVER_PORT")
}

func GetExpensecategoryrelationServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "EXPENSECATEGORYRELATION_SERVER_PORT")
}

// GetCurrencyServerPort returns the port the expense service will run on
func GetCurrencyServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "CURRENCY_SERVER_PORT")
}

func GetDbUser(ctx context.Context) string {
	return MustLookupString(ctx, "DB_USER")
}

func GetDbPassword(ctx context.Context) string {
	return MustLookupString(ctx, "DB_PASSWORD")
}

func GetDbHost(ctx context.Context) string {
	return MustLookupString(ctx, "DB_HOST")
}

func GetDbPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "DB_PORT")
}

func GetDbName(ctx context.Context) string {
	return MustLookupString(ctx, "DB_NAME")
}

// GetServerPort returns the port the service will run on
func GetReflectionServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "REFLECTION_SERVER_PORT")
}

// GetGlobalDomain returns the infrastructure's global domain which can be used for various purposes like error details
func GetGlobalDomain(ctx context.Context) string {
	return MustLookupString(ctx, "GLOBAL_DOMAIN")
}

// GetDBSelectErrorReason returns the error reason that a DB select to the database failed in UPPER_SNAKE_CASE
func GetDBSelectErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "DB_SELECT_ERROR_REASON")
}

// GetDBInsertErrorReason returns the error reason that a DB insert to the database failed in UPPER_SNAKE_CASE
func GetDBInsertErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "DB_INSERT_ERROR_REASON")
}

// GetDBDeleteErrorReason returns the error reason that a DB delete to the database failed in UPPER_SNAKE_CASE
func GetDBDeleteErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "DB_DELETE_ERROR_REASON")
}

// GetDBUpdateErrorReason returns the error reason that a DB update to the database failed in UPPER_SNAKE_CASE
func GetDBUpdateErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "DB_UPDATE_ERROR_REASON")
}

// GetMessagePublicationErrorReason returns the error reason that a message could not be published in UPPER_SNAKE_CASE
func GetMessagePublicationErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "MESSAGE_PUBLICATION_ERROR_REASON")
}

// GetMessageSubscriptionErrorReason returns the error reason that a subscription to messages failed in UPPER_SNAKE_CASE
func GetMessageSubscriptionErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "MESSAGE_SUBSCRIPTION_ERROR_REASON")
}

// GetSendCurrentResourceErrorReason returns the error reason that the current resource could not be sent in UPPER_SNAKE_CASE
func GetSendCurrentResourceErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "SEND_CURRENT_RESOURCE_ERROR_REASON")
}

// GetSendStreamAliveErrorReason returns the error reason that an alive message could not be sent in UPPER_SNAKE_CASE
func GetSendStreamAliveErrorReason(ctx context.Context) string {
	return MustLookupString(ctx, "SEND_STREAM_ALIVE_ERROR_REASON")
}

// GetNatsServerHost returns the host address of the NATS server
func GetNatsServerHost(ctx context.Context) string {
	return MustLookupString(ctx, "NATS_SERVER_HOST")
}

// GetNatsServerPort returns the host port of the NATS server
func GetNatsServerPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "NATS_SERVER_PORT")
}

func GetTraceCollectorHost(ctx context.Context) string {
	return MustLookupString(ctx, "TRACE_COLLECTOR_HOST")
}

func GetTraceCollectorPort(ctx context.Context) uint16 {
	return MustLookupUint16(ctx, "TRACE_COLLECTOR_PORT")
}

// TODO: as env variable
// GetExpenseSplitterSubject returns the name of the root subject all expense splitter events are published under
func GetExpenseSplitterSubject() string {
	return "expensesplitter"
}

// TODO: as env variable with %s parameter
// GetGroupCreatedSubject returns the name of the subject events are published on when a group was created
func GetGroupCreatedSubject(groupId string) string {
	return fmt.Sprintf("%s.created", GetGroupSubject(groupId))
}

// TODO: as env variable with %s parameter
// GetGroupDeletedSubject returns the name of the subject events are published on when a group was deleted
func GetGroupDeletedSubject(groupId string) string {
	return fmt.Sprintf("%s.deleted", GetGroupSubject(groupId))
}

// TODO: as env variable with %s parameter
// GetGroupUpdatedSubject returns the name of the subject events are published on when a group was updated
func GetGroupUpdatedSubject(groupId string) string {
	return fmt.Sprintf("%s.updated", GetGroupSubject(groupId))
}

// TODO: as env variable with %s parameter
// GetGroupSubject returns the name of the subject events of a single group are published on
func GetGroupSubject(groupId string) string {
	return fmt.Sprintf("%s.%s", GetGroupsSubject(), groupId)
}

// TODO: as env variable
// GetGroupsSubject returns the name of the subject events of all groups are published on
func GetGroupsSubject() string {
	return fmt.Sprintf("%s.group", GetExpenseSplitterSubject())
}

func GetGroupSourceStreamName() string {
	return "EXPENSESPLITTER_GROUP"
}

// TODO: as env variable with %s parameter
// GetPersonCreatedSubject returns the name of the subject events are published on when a person was created
func GetPersonCreatedSubject(groupId string, personId string) string {
	return fmt.Sprintf("%s.created", GetPersonSubject(groupId, personId))
}

// TODO: as env variable with %s parameter
// GetPersonDeletedSubject returns the name of the subject events are published on when a person was deleted
func GetPersonDeletedSubject(groupId string, personId string) string {
	return fmt.Sprintf("%s.deleted", GetPersonSubject(groupId, personId))
}

// TODO: as env variable with %s parameter
// GetPersonUpdatedSubject returns the name of the subject events are published on when a person was updated
func GetPersonUpdatedSubject(groupId string, personId string) string {
	return fmt.Sprintf("%s.updated", GetPersonSubject(groupId, personId))
}

// TODO: as env variable with %s parameter
// GetPersonSubject returns the name of the subject events of a single person are published on
func GetPersonSubject(groupId string, personId string) string {
	return fmt.Sprintf("%s.%s", GetPeopleSubject(groupId), personId)
}

// TODO: as env variable
// GetPeopleSubject returns the name of the subject events of all people are published on
func GetPeopleSubject(groupId string) string {
	return fmt.Sprintf("%s.person", GetGroupSubject(groupId))
}

func GetPersonSourceStreamName() string {
	return "EXPENSESPLITTER_PERSON"
}

// TODO: as env variable with %s parameter
// GetCategoryCreatedSubject returns the name of the subject events are published on when a category was created
func GetCategoryCreatedSubject(groupId string, categoryId string) string {
	return fmt.Sprintf("%s.created", GetCategorySubject(groupId, categoryId))
}

// TODO: as env variable with %s parameter
// GetCategoryDeletedSubject returns the name of the subject events are published on when a category was deleted
func GetCategoryDeletedSubject(groupId string, categoryId string) string {
	return fmt.Sprintf("%s.deleted", GetCategorySubject(groupId, categoryId))
}

// TODO: as env variable with %s parameter
// GetCategoryUpdatedSubject returns the name of the subject events are published on when a category was updated
func GetCategoryUpdatedSubject(groupId string, categoryId string) string {
	return fmt.Sprintf("%s.updated", GetCategorySubject(groupId, categoryId))
}

// TODO: as env variable with %s parameter
// GetCategorySubject returns the name of the subject events of a single category are published on
func GetCategorySubject(groupId string, categoryId string) string {
	return fmt.Sprintf("%s.%s", GetCategoriesSubject(groupId), categoryId)
}

// TODO: as env variable
// GetCategoriesSubject returns the name of the subject events of all categories are published on
func GetCategoriesSubject(groupId string) string {
	return fmt.Sprintf("%s.category", GetGroupSubject(groupId))
}

func GetCategorySourceStreamName() string {
	return "EXPENSESPLITTER_CATEGORY"
}

// TODO: as env variable with %s parameter
// GetExpenseCreatedSubject returns the name of the subject events are published on when a expense was created
func GetExpenseCreatedSubject(groupId string, expenseId string) string {
	return fmt.Sprintf("%s.created", GetExpenseSubject(groupId, expenseId))
}

// TODO: as env variable with %s parameter
// GetExpenseDeletedSubject returns the name of the subject events are published on when a expense was deleted
func GetExpenseDeletedSubject(groupId string, expenseId string) string {
	return fmt.Sprintf("%s.deleted", GetExpenseSubject(groupId, expenseId))
}

// TODO: as env variable with %s parameter
// GetExpenseUpdatedSubject returns the name of the subject events are published on when a expense was updated
func GetExpenseUpdatedSubject(groupId string, expenseId string) string {
	return fmt.Sprintf("%s.updated", GetExpenseSubject(groupId, expenseId))
}

// TODO: as env variable with %s parameter
// GetExpenseSubject returns the name of the subject events of a single expense are published on
func GetExpenseSubject(groupId string, expenseId string) string {
	return fmt.Sprintf("%s.%s", GetExpensesSubject(groupId), expenseId)
}

// TODO: as env variable
// GetExpensesSubject returns the name of the subject events of all expenses are published on
func GetExpensesSubject(groupId string) string {
	return fmt.Sprintf("%s.expense", GetGroupSubject(groupId))
}

func GetExpenseSourceStreamName() string {
	return "EXPENSESPLITTER_EXPENSE"
}

// TODO: as env variable with %s parameter
// GetExpenseStakeCreatedSubject returns the name of the subject events are published on when a expense stake was created
func GetExpenseStakeCreatedSubject(groupId string, expenseId string, stakeId string) string {
	return fmt.Sprintf("%s.created", GetExpenseStakeSubject(groupId, expenseId, stakeId))
}

// TODO: as env variable with %s parameter
// GetExpenseStakeDeletedSubject returns the name of the subject events are published on when a expense stake was deleted
func GetExpenseStakeDeletedSubject(groupId string, expenseId string, stakeId string) string {
	return fmt.Sprintf("%s.deleted", GetExpenseStakeSubject(groupId, expenseId, stakeId))
}

// TODO: as env variable with %s parameter
// GetExpenseStakeUpdatedSubject returns the name of the subject events are published on when a expense stake was updated
func GetExpenseStakeUpdatedSubject(groupId string, expenseId string, stakeId string) string {
	return fmt.Sprintf("%s.updated", GetExpenseStakeSubject(groupId, expenseId, stakeId))
}

// TODO: as env variable with %s parameter
// GetExpenseStakeSubject returns the name of the subject events of a single expense stake are published on
func GetExpenseStakeSubject(groupId string, expenseId string, stakeId string) string {
	return fmt.Sprintf("%s.%s", GetExpenseStakesSubject(groupId, expenseId), stakeId)
}

// TODO: as env variable
// GetExpenseStakesSubject returns the name of the subject events of all expense stakes are published on
func GetExpenseStakesSubject(groupId string, expenseId string) string {
	return fmt.Sprintf("%s.expensestake", GetExpenseSubject(groupId, expenseId))
}

func GetExpenseStakeSourceStreamName() string {
	return "EXPENSESPLITTER_EXPENSESTAKE"
}

// TODO: as env variable with %s parameter
// GetCurrencyCreatedSubject returns the name of the subject events are published on when a group was created
func GetCurrencyCreatedSubject(currencyId string) string {
	return fmt.Sprintf("%s.created", GetCurrencySubject(currencyId))
}

// TODO: as env variable with %s parameter
// GetCurrencyDeletedSubject returns the name of the subject events are published on when a group was deleted
func GetCurrencyDeletedSubject(currencyId string) string {
	return fmt.Sprintf("%s.deleted", GetCurrencySubject(currencyId))
}

// TODO: as env variable with %s parameter
// GetCurrencyUpdatedSubject returns the name of the subject events are published on when a group was updated
func GetCurrencyUpdatedSubject(groupId string) string {
	return fmt.Sprintf("%s.updated", GetCurrencySubject(groupId))
}

// TODO: as env variable with %s parameter
// GetCurrencySubject returns the name of the subject events of a single group are published on
func GetCurrencySubject(currencyId string) string {
	return fmt.Sprintf("%s.%s", GetCurrenciesSubject(), currencyId)
}

// TODO: as env variable
// GetCurrenciesSubject returns the name of the subject events of all groups are published on
func GetCurrenciesSubject() string {
	return fmt.Sprintf("%s.currency", GetExpenseSplitterSubject())
}

func GetCurrencySourceStreamName() string {
	return "EXPENSESPLITTER_CURRENCY"
}

// TODO: as env variable with %s parameter
// GetExpenseCategoryRelationCreatedSubject returns the name of the subject events are published on when a expense stake was created
func GetExpenseCategoryRelationCreatedSubject(groupId string, expenseId string, categoryId string) string {
	return fmt.Sprintf("%s.created", GetExpenseCategoryRelationSubject(groupId, expenseId, categoryId))
}

// TODO: as env variable with %s parameter
// GetExpenseCategoryRelationDeletedSubject returns the name of the subject events are published on when a expense stake was deleted
func GetExpenseCategoryRelationDeletedSubject(groupId string, expenseId string, categoryId string) string {
	return fmt.Sprintf("%s.deleted", GetExpenseCategoryRelationSubject(groupId, expenseId, categoryId))
}

// TODO: as env variable with %s parameter
// GetExpenseCategoryRelationSubject returns the name of the subject events of a single expense stake are published on
func GetExpenseCategoryRelationSubject(groupId string, expenseId string, categoryId string) string {
	return fmt.Sprintf("%s.%s", GetExpenseCategoryRelationsSubject(groupId, expenseId), categoryId)
}

// TODO: as env variable
// GetExpenseCategoryRelationsSubject returns the name of the subject events of all expense stakes are published on
func GetExpenseCategoryRelationsSubject(groupId string, expenseId string) string {
	return fmt.Sprintf("%s.category", GetExpenseSubject(groupId, expenseId))
}

func GetExpenseCategoryRelationSourceStreamName() string {
	return "EXPENSESPLITTER_EXPENSECATEGORYRELATION"
}

// TODO: as env variable
// GetHttpStatusCodeKey returns the header key used internally to modify the http status code as suggested here: https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/
func GetHttpStatusCodeKey() string {
	return "x-http-code"
}
