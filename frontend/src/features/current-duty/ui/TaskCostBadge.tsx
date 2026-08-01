import { Badge } from '@mantine/core'

type Props = {
    cost: number
}

export function TaskCostBadge({ cost }: Props) {
    return (
        <Badge
            radius="sm"
            variant="filled"
            styles={{
                root: {
                    backgroundColor: '#EEF2F1',
                    color: 'var(--app-color-text)',
                    fontWeight: 500,
                    textTransform: 'none',
                },
                label: {
                    textTransform: 'none',
                },
            }}
        >
            {cost} баллов
        </Badge>
    )
}
