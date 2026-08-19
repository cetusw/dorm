import {
    ListIcon,
    UserListIcon,
    UsersIcon,
} from '@phosphor-icons/react'
import { Box, Text, UnstyledButton } from '@mantine/core'

import type { DutyTaskSelect } from '../model/types'
import classes from './MobileDutyBottomNav.module.css'

type Props = {
    activeSelect: DutyTaskSelect
    onChange: (select: DutyTaskSelect) => void
}

const items: Array<{
    icon: typeof UserListIcon
    label: string
    value: DutyTaskSelect
}> = [
    { icon: UserListIcon, label: 'Мои задачи', value: 'mine' },
    { icon: ListIcon, label: 'Задачи', value: 'all' },
    { icon: UsersIcon, label: 'Команда', value: 'team' },
]

export function MobileDutyBottomNav({ activeSelect, onChange }: Props) {
    const activeIndex = Math.max(items.findIndex((item) => item.value === activeSelect), 0)

    return (
        <Box className={classes.panel} hiddenFrom="md">
            <Box className={classes.items}>
                <Box
                    aria-hidden="true"
                    className={classes.indicator}
                    style={{
                        transform: `translateX(calc(${activeIndex} * (100% + 6px)))`,
                    }}
                />
                {items.map((item) => {
                    const Icon = item.icon

                    return (
                        <Box
                            className={classes.item}
                            key={item.value}
                        >
                            <UnstyledButton
                                className={classes.button}
                                aria-pressed={activeSelect === item.value}
                                onClick={() => onChange(item.value)}
                            >
                                <Icon size={24} weight="regular" />
                                <Text className={classes.label}>
                                    {item.label}
                                </Text>
                            </UnstyledButton>
                        </Box>
                    )
                })}
            </Box>
        </Box>
    )
}
