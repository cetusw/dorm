import {Group, Select} from '@mantine/core'

import type {DutyTaskSelect, ResidentDutyGroupOption,} from '../model/types'

type Props = {
    activeSelect: DutyTaskSelect
    groups: ResidentDutyGroupOption[]
    selectedGroupId: string
    showGroupSelect?: boolean
    visibleSelects: DutyTaskSelect[]
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
    {label: 'Задачи на проверку', value: 'review'},
]

export function DutyTaskSelects({
                                    activeSelect,
                                    groups,
                                    selectedGroupId,
                                    showGroupSelect = true,
                                    visibleSelects,
                                    onGroupChange,
                                    onChange,
                                }: Props) {
    const visibleSelectOptions = selectOptions.filter((selectOption) =>
        visibleSelects.includes(selectOption.value),
    )

    return (
        <Group
            gap="sm"
            wrap="nowrap"
            align="flex-end"
        >
            {showGroupSelect && (
                <Select
                    aria-label="Группа"
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
                    style={{
                        flex: 1,
                        minWidth: 0,
                    }}
                />
            )}

            {visibleSelectOptions.length > 0 && (
                <Select
                    aria-label="Раздел задач"
                    data={visibleSelectOptions}
                    value={activeSelect}
                    onChange={(value) => {
                        if (value) {
                            onChange(value as DutyTaskSelect)
                        }
                    }}
                    allowDeselect={false}
                    style={{
                        flex: 1,
                        minWidth: 0,
                    }}
                />
            )}
        </Group>
    )
}
