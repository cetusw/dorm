import type { ReactNode } from 'react'

import { Box, Group, SegmentedControl, Stack } from '@mantine/core'

import type { DutyTaskSelect } from '../model/types'
import classes from './SegmentedControl.module.css'

type Props = {
    activeSelect: DutyTaskSelect
    visibleSelects: DutyTaskSelect[]
    rightSection?: ReactNode
    onChange: (select: DutyTaskSelect) => void
}

const selectOptions: Array<{
    label: string
    value: DutyTaskSelect
}> = [
    {label: 'Мои задачи', value: 'mine'},
    {label: 'Свободные задачи', value: 'free'},
    {label: 'Все задачи', value: 'all'},
    {label: 'Задачи на проверке', value: 'verification'},
    {label: 'Команда', value: 'team'},
]

export function DutyTaskSelects({
    activeSelect,
    visibleSelects,
    rightSection,
    onChange,
}: Props) {
    const visibleSelectOptions = selectOptions.filter((selectOption) =>
        visibleSelects.includes(selectOption.value),
    )

    return (
        <Stack gap="sm">
            {(visibleSelectOptions.length > 0 || rightSection) && (
                <Group align="center" justify="space-between" gap="sm" wrap="nowrap">
                    {visibleSelectOptions.length > 0 && (
                        <Box
                            className={classes.scrollArea}
                            style={{ flex: 1, minWidth: 0, maxWidth: '100%' }}
                        >
                            <SegmentedControl
                                aria-label="Раздел задач"
                                data={visibleSelectOptions}
                                value={activeSelect}
                                classNames={{
                                    control: classes.control,
                                    root: classes.root,
                                    label: classes.label,
                                }}
                                onChange={(value) => onChange(value as DutyTaskSelect)}
                            />
                        </Box>
                    )}

                    {rightSection && (
                        <Box visibleFrom="md" style={{ flexShrink: 0 }}>
                            {rightSection}
                        </Box>
                    )}
                </Group>
            )}
        </Stack>
    )
}
