import type { ReactNode } from 'react'

import { Box, Group, SegmentedControl, Select, Stack } from '@mantine/core'

import type {DutyTaskSelect, ResidentDutyGroupOption,} from '../model/types'
import classes from './SegmentedControl.module.css'

type Props = {
    activeSelect: DutyTaskSelect
    groups: ResidentDutyGroupOption[]
    selectedGroupId: string
    showGroupSelect?: boolean
    visibleSelects: DutyTaskSelect[]
    rightSection?: ReactNode
    onGroupChange: (groupId: string) => void
    onChange: (select: DutyTaskSelect) => void
}

const selectOptions: Array<{
    label: string
    value: DutyTaskSelect
}> = [
    {label: 'Мои задачи', value: 'mine'},
    {label: 'Свободные задачи', value: 'free'},
    {label: 'Все задачи', value: 'all'},
    {label: 'Команда', value: 'team'},
    {label: 'Задачи на проверку', value: 'review'},
]

export function DutyTaskSelects({
    activeSelect,
    groups,
    selectedGroupId,
    showGroupSelect = true,
    visibleSelects,
    rightSection,
    onGroupChange,
    onChange,
}: Props) {
    const visibleSelectOptions = selectOptions.filter((selectOption) =>
        visibleSelects.includes(selectOption.value),
    )

    return (
        <Stack gap="sm">
            {showGroupSelect && (
                <Select
                    aria-label="Группа"
                    autoComplete="off"
                    data={groups.map((group) => ({
                        value: group.id,
                        label: group.name,
                    }))}
                    value={selectedGroupId}
                    onChange={(value) => {
                        if (value) {
                            onGroupChange(value)
                        }
                    }}
                    allowDeselect={false}
                    style={{ width: '100%' }}
                />
            )}

            {(visibleSelectOptions.length > 0 || rightSection) && (
                <Group align="center" justify="space-between" gap="sm" wrap="wrap">
                    {visibleSelectOptions.length > 0 && (
                        <Box style={{ minWidth: 0, maxWidth: '100%' }}>
                            <SegmentedControl
                                aria-label="Раздел задач"
                                data={visibleSelectOptions}
                                value={activeSelect}
                                classNames={{
                                    root: classes.root,
                                    label: classes.label,
                                }}
                                onChange={(value) => onChange(value as DutyTaskSelect)}
                            />
                        </Box>
                    )}

                    {rightSection}
                </Group>
            )}
        </Stack>
    )
}
