import { Button, Group, Paper, ScrollArea, Table, Text } from '@mantine/core'

import type { DormitoryListItem } from '../model/types'

type Props = {
    dormitories: DormitoryListItem[]
}

export function DormitoriesTable({ dormitories }: Props) {
    return (
        <Paper withBorder radius="lg" bg="white">
            <ScrollArea>
                <Table horizontalSpacing="lg" verticalSpacing="md" highlightOnHover miw={920}>
                    <Table.Thead>
                        <Table.Tr>
                            <Table.Th w="26%">
                                <Text size="sm" fw={700}>
                                    Название
                                </Text>
                            </Table.Th>
                            <Table.Th w="34%">
                                <Text size="sm" fw={700}>
                                    Адрес
                                </Text>
                            </Table.Th>
                            <Table.Th w="22%">
                                <Text size="sm" fw={700}>
                                    Глава
                                </Text>
                            </Table.Th>
                            <Table.Th w="18%">
                                <Text size="sm" fw={700}>
                                    Действия
                                </Text>
                            </Table.Th>
                        </Table.Tr>
                    </Table.Thead>
                    <Table.Tbody>
                        {dormitories.map((dormitory) => (
                            <Table.Tr key={dormitory.id}>
                                <Table.Td>
                                    <Text fw={600}>{dormitory.name}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text>{dormitory.address}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text c={dormitory.leader ? undefined : 'dimmed'}>
                                        {dormitory.leader?.name ?? 'Не назначен'}
                                    </Text>
                                </Table.Td>
                                <Table.Td>
                                    <Group gap="xs" wrap="nowrap">
                                        <Button variant="default" size="compact-sm" radius="md">
                                            Редактировать
                                        </Button>
                                        <Button variant="default" color="red" size="compact-sm" radius="md">
                                            Удалить
                                        </Button>
                                    </Group>
                                </Table.Td>
                            </Table.Tr>
                        ))}
                    </Table.Tbody>
                </Table>
            </ScrollArea>
        </Paper>
    )
}
