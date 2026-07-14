import { Alert, Button, Center, Loader } from '@mantine/core'

import { useDormitories } from '../../features/dormitories/model/useDormitories'
import { DormitoriesTable } from '../../features/dormitories/ui/DormitoriesTable'
import { PageFrame } from '../../shared/ui/PageFrame'

export function DormitoriesPage() {
    const { dormitories, loading, error } = useDormitories()

    const titleActions = (
        <Button radius="md">+ Создать общежитие</Button>
    )

    if (loading) {
        return (
            <PageFrame title="Общежития" titleActions={titleActions}>
                <Center py="xl">
                    <Loader />
                </Center>
            </PageFrame>
        )
    }

    if (error) {
        return (
            <PageFrame title="Общежития" titleActions={titleActions}>
                <Alert color="red" title="Ошибка">
                    {error}
                </Alert>
            </PageFrame>
        )
    }

    if (dormitories.length === 0) {
        return (
            <PageFrame title="Общежития" titleActions={titleActions}>
                <Alert color="gray">
                    Общежития пока не добавлены.
                </Alert>
            </PageFrame>
        )
    }

    return (
        <PageFrame title="Общежития" titleActions={titleActions}>
            <DormitoriesTable dormitories={dormitories} />
        </PageFrame>
    )
}
