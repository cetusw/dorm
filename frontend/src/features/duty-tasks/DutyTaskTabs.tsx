import { Button, Group } from '@mantine/core'

export type DutyTaskTab = 'mine' | 'free' | 'all' | 'review'

type Props = {
    activeTab: DutyTaskTab
    onChange: (tab: DutyTaskTab) => void
}

const tabs: Array<{ label: string; value: DutyTaskTab }> = [
    { label: 'Мои задачи', value: 'mine' },
    { label: 'Свободные задачи', value: 'free' },
    { label: 'Все задачи', value: 'all' },
    { label: 'Задачи на проверку', value: 'review' },
]

export function DutyTaskTabs({ activeTab, onChange }: Props) {
    return (
        <Group gap="sm" wrap="wrap">
            {tabs.map((tab) => {
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
