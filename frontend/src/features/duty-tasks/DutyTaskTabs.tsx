import { Button, Group, Select } from '@mantine/core'

import type { ResidentDutyGroupOption } from './types'

export type DutyTaskTab = 'mine' | 'free' | 'all' | 'review'

type Props = {
    activeTab: DutyTaskTab
    groups: ResidentDutyGroupOption[]
    selectedGroupId: string
    showGroupSelect?: boolean
    visibleTabs: DutyTaskTab[]
    onGroupChange: (groupId: string) => void
    onChange: (tab: DutyTaskTab) => void
}

const tabs: Array<{ label: string; value: DutyTaskTab }> = [
    { label: 'Мои задачи', value: 'mine' },
    { label: 'Свободные задачи', value: 'free' },
    { label: 'Все задачи', value: 'all' },
    { label: 'Задачи на проверку', value: 'review' },
]

export function DutyTaskTabs({
    activeTab,
    groups,
    selectedGroupId,
    showGroupSelect = true,
    visibleTabs,
    onGroupChange,
    onChange,
}: Props) {
    return (
        <Group gap="sm" wrap="wrap">
            {showGroupSelect && (
                <Select
                    aria-label="Группа"
                    placeholder="Группа"
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
                    w={{ base: '100%', sm: 240 }}
                />
            )}
            {tabs.filter((tab) => visibleTabs.includes(tab.value)).map((tab) => {
                const isActive = tab.value === activeTab

                return (
                    <Button
                        key={tab.value}
                        variant="default"
                        radius="md"
                        color="gray"
                        onClick={() => onChange(tab.value)}
                        styles={(theme) => ({
                            root: {
                                backgroundColor: isActive ? theme.colors.gray[2] : theme.white,
                                borderColor: theme.colors.gray[3],
                                color: theme.black,
                                '&:hover': {
                                    backgroundColor: theme.colors.gray[2],
                                },
                            },
                        })}
                    >
                        {tab.label}
                    </Button>
                )
            })}
        </Group>
    )
}
