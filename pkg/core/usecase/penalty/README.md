# Лабораторная работа №6

## Пояснение к классу
`PenaltyService` отвечает за работу с выдачей штрафов жителям общежития. 
При выполнении действий со штрафами сначала должны пройти проверки бизнес-правил, например, может ли текущий пользователь выдавать штрафы, кому он может выдавать штрафы и так далее.

## Запуск тестов
```bash
go test ./pkg/core/usecase/penalty -v
```

Результат:
```bash
=== RUN   TestListResidents_DormitoryLeaderPassesScopeAndResponse
--- PASS: TestListResidents_DormitoryLeaderPassesScopeAndResponse (0.00s)
=== RUN   TestSearchResidents_GroupLeaderScopeAndSearch
--- PASS: TestSearchResidents_GroupLeaderScopeAndSearch (0.00s)
=== RUN   TestListResidents_CombinesScopesWithoutDuplicates
--- PASS: TestListResidents_CombinesScopesWithoutDuplicates (0.00s)
=== RUN   TestPenaltyScope_DeniesOrdinaryAndMissingUser
=== RUN   TestPenaltyScope_DeniesOrdinaryAndMissingUser/ordinary
=== RUN   TestPenaltyScope_DeniesOrdinaryAndMissingUser/missing
--- PASS: TestPenaltyScope_DeniesOrdinaryAndMissingUser (0.00s)
    --- PASS: TestPenaltyScope_DeniesOrdinaryAndMissingUser/ordinary (0.00s)
    --- PASS: TestPenaltyScope_DeniesOrdinaryAndMissingUser/missing (0.00s)
=== RUN   TestListAndSearch_WrapQueryErrors
=== RUN   TestListAndSearch_WrapQueryErrors/list
=== RUN   TestListAndSearch_WrapQueryErrors/search
--- PASS: TestListAndSearch_WrapQueryErrors (0.00s)
    --- PASS: TestListAndSearch_WrapQueryErrors/list (0.00s)
    --- PASS: TestListAndSearch_WrapQueryErrors/search (0.00s)
=== RUN   TestGetResidentPenalties_ReturnsHistoryBalanceAndName
--- PASS: TestGetResidentPenalties_ReturnsHistoryBalanceAndName (0.00s)
=== RUN   TestGetResidentPenalties_RejectsMissingAndOutOfScopeResident
=== RUN   TestGetResidentPenalties_RejectsMissingAndOutOfScopeResident/missing
=== RUN   TestGetResidentPenalties_RejectsMissingAndOutOfScopeResident/outside_scope
--- PASS: TestGetResidentPenalties_RejectsMissingAndOutOfScopeResident (0.00s)
    --- PASS: TestGetResidentPenalties_RejectsMissingAndOutOfScopeResident/missing (0.00s)
    --- PASS: TestGetResidentPenalties_RejectsMissingAndOutOfScopeResident/outside_scope (0.00s)
=== RUN   TestGetCurrentUserPenalties_ReturnsOwnHistoryAndErrors
--- PASS: TestGetCurrentUserPenalties_ReturnsOwnHistoryAndErrors (0.00s)
=== RUN   TestCreatePenalty_NormalizesInputCreatesIssueAtControlledTime
--- PASS: TestCreatePenalty_NormalizesInputCreatesIssueAtControlledTime (0.00s)
=== RUN   TestCreatePenalty_RejectsInvalidInputsAndRepositoryError
=== RUN   TestCreatePenalty_RejectsInvalidInputsAndRepositoryError/bad_UUID
=== RUN   TestCreatePenalty_RejectsInvalidInputsAndRepositoryError/bad_weight
=== RUN   TestCreatePenalty_RejectsInvalidInputsAndRepositoryError/empty_reason
--- PASS: TestCreatePenalty_RejectsInvalidInputsAndRepositoryError (0.00s)
    --- PASS: TestCreatePenalty_RejectsInvalidInputsAndRepositoryError/bad_UUID (0.00s)
    --- PASS: TestCreatePenalty_RejectsInvalidInputsAndRepositoryError/bad_weight (0.00s)
    --- PASS: TestCreatePenalty_RejectsInvalidInputsAndRepositoryError/empty_reason (0.00s)
=== RUN   TestResolvePenalty_NormalizesAndUsesAtomicRepository
--- PASS: TestResolvePenalty_NormalizesAndUsesAtomicRepository (0.00s)
=== RUN   TestResolvePenalty_PropagatesRepositoryBalanceAndReservationConflicts
=== RUN   TestResolvePenalty_PropagatesRepositoryBalanceAndReservationConflicts/penalty_balance_is_insufficient
=== RUN   TestResolvePenalty_PropagatesRepositoryBalanceAndReservationConflicts/penalty_balance_is_reserved_by_individual_tasks
--- PASS: TestResolvePenalty_PropagatesRepositoryBalanceAndReservationConflicts (0.00s)
    --- PASS: TestResolvePenalty_PropagatesRepositoryBalanceAndReservationConflicts/penalty_balance_is_insufficient (0.00s)
    --- PASS: TestResolvePenalty_PropagatesRepositoryBalanceAndReservationConflicts/penalty_balance_is_reserved_by_individual_tasks (0.00s)
=== RUN   TestDeletePenaltyEntry_DeletesAccessibleEntry
--- PASS: TestDeletePenaltyEntry_DeletesAccessibleEntry (0.00s)
=== RUN   TestDeletePenaltyEntry_HandlesMissingForbiddenAndRepositoryConflict
=== RUN   TestDeletePenaltyEntry_HandlesMissingForbiddenAndRepositoryConflict/missing
=== RUN   TestDeletePenaltyEntry_HandlesMissingForbiddenAndRepositoryConflict/forbidden
=== RUN   TestDeletePenaltyEntry_HandlesMissingForbiddenAndRepositoryConflict/reserved
--- PASS: TestDeletePenaltyEntry_HandlesMissingForbiddenAndRepositoryConflict (0.00s)
    --- PASS: TestDeletePenaltyEntry_HandlesMissingForbiddenAndRepositoryConflict/missing (0.00s)
    --- PASS: TestDeletePenaltyEntry_HandlesMissingForbiddenAndRepositoryConflict/forbidden (0.00s)
    --- PASS: TestDeletePenaltyEntry_HandlesMissingForbiddenAndRepositoryConflict/reserved (0.00s)
=== RUN   TestUpdatePenaltyEntry_PreservesTypeAndTimeAndPropagatesInvariantError
--- PASS: TestUpdatePenaltyEntry_PreservesTypeAndTimeAndPropagatesInvariantError (0.00s)
=== RUN   TestUpdatePenaltyEntry_ReportsMissingEntry
--- PASS: TestUpdatePenaltyEntry_ReportsMissingEntry (0.00s)
=== RUN   TestScope_WrapsCollaboratorFailures
--- PASS: TestScope_WrapsCollaboratorFailures (0.00s)
=== RUN   TestResidentAndEntryLookupErrorsAreWrapped
--- PASS: TestResidentAndEntryLookupErrorsAreWrapped (0.00s)
=== RUN   TestGetResidentPenalties_WrapsHistoryQueryError
--- PASS: TestGetResidentPenalties_WrapsHistoryQueryError (0.00s)
=== RUN   TestResolvePenalty_RejectsInvalidInputBeforeWrite
--- PASS: TestResolvePenalty_RejectsInvalidInputBeforeWrite (0.00s)
=== RUN   TestCurrentUserAndResidentValidationErrors
--- PASS: TestCurrentUserAndResidentValidationErrors (0.00s)
=== RUN   TestCreateAndResolve_RejectMissingResidentAndInvalidWeight
--- PASS: TestCreateAndResolve_RejectMissingResidentAndInvalidWeight (0.00s)
=== RUN   TestUpdatePenaltyEntry_RejectsInvalidMutableValues
--- PASS: TestUpdatePenaltyEntry_RejectsInvalidMutableValues (0.00s)
PASS
ok      dorm/pkg/core/usecase/penalty   0.026s
```

## Покрытие класса
```bash
go test ./pkg/core/usecase/penalty -cover
```

Результат: 
```bash
ok      dorm/pkg/core/usecase/penalty   0.024s  coverage: 90.6% of statements
```
