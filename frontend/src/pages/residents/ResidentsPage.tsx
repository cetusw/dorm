import { Alert, Button, Center, Loader } from '@mantine/core'

import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useResidents } from '../../features/residents/model/useResidents'
import { ResidentsTable } from '../../features/residents/ui/ResidentsTable'
import { ManagementPageFrame } from '../../shared/ui/ManagementPageFrame'

export function ResidentsPage() {
    const selectedDormitoryId = useSelectedDormitoryId()
    const { residents, loading, error } = useResidents(selectedDormitoryId)

    const titleActions = (
        <Button radius="md">
            + Создать жителя
        </Button>
    )

    if (!selectedDormitoryId) {
        return (
            <ManagementPageFrame title="Жители" titleActions={titleActions}>
                <Alert color="gray">
                    Выберите общежитие в верхней панели.
                </Alert>
            </ManagementPageFrame>
        )
    }

    if (loading) {
        return (
            <ManagementPageFrame title="Жители" titleActions={titleActions}>
                <Center py="xl">
                    <Loader />
                </Center>
            </ManagementPageFrame>
        )
    }

    if (error) {
        return (
            <ManagementPageFrame title="Жители" titleActions={titleActions}>
                <Alert color="red" title="Ошибка">
                    {error}
                </Alert>
            </ManagementPageFrame>
        )
    }

    if (residents.length === 0) {
        return (
            <ManagementPageFrame title="Жители" titleActions={titleActions}>
                <Alert color="gray">
                    В выбранном общежитии пока нет жителей.
                </Alert>
            </ManagementPageFrame>
        )
    }

    return (
        <ManagementPageFrame title="Жители" titleActions={titleActions}>
            <ResidentsTable residents={residents} />
        </ManagementPageFrame>
    )
}
