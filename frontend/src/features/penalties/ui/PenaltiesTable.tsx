import { Table, Text } from '@mantine/core'

import type { PenaltyResidentSummary } from '../model/types'
import { formatPenaltyWeight } from '../model/utils'
import { AppTable, appTableClasses } from '../../../shared/ui/AppTable'
import classes from './PenaltiesTable.module.css'

type Props = {
    residents: PenaltyResidentSummary[]
    onOpenResident: (residentId: string) => void
}

export function PenaltiesTable({ residents, onOpenResident }: Props) {
    return (
        <AppTable minWidth={520}>
            <Table.Thead>
                <Table.Tr>
                    <Table.Th>Имя</Table.Th>
                    <Table.Th w={180}>Предупреждения</Table.Th>
                </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
                {residents.map((resident) => (
                    <Table.Tr
                        key={resident.user_id}
                        className={`${appTableClasses.interactiveRow} ${appTableClasses.compactRow} ${resident.threshold_reached ? classes.thresholdRow : ''}`}
                        tabIndex={0}
                        aria-label={`Открыть предупреждения жителя ${resident.full_name}`}
                        onClick={() => onOpenResident(resident.user_id)}
                        onKeyDown={(event) => {
                            if (event.key === 'Enter' || event.key === ' ') {
                                event.preventDefault()
                                onOpenResident(resident.user_id)
                            }
                        }}
                    >
                        <Table.Td>
                            <Text fw={500} inherit>{resident.full_name}</Text>
                        </Table.Td>
                        <Table.Td>
                            <Text inherit>{formatPenaltyWeight(resident.total_weight)}</Text>
                        </Table.Td>
                    </Table.Tr>
                ))}
            </Table.Tbody>
        </AppTable>
    )
}
