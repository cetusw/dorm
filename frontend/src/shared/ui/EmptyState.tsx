import { MagnifyingGlassIcon } from '@phosphor-icons/react'
import { Stack, Text, ThemeIcon, Title } from '@mantine/core'

type Props = {
    title: string
    description?: string
}

export function EmptyState({ title, description }: Props) {
    return (
        <Stack
            align="center"
            gap="md"
            py={48}
            px="md"
            ta="center"
        >
            <ThemeIcon
                size={64}
                radius="xl"
                variant="light"
                color="gray"
            >
                <MagnifyingGlassIcon size={32} />
            </ThemeIcon>

            <Stack gap={6} align="center">
                <Title order={3} fw={600}>
                    {title}
                </Title>

                {description ? (
                    <Text c="dimmed" size="sm" maw={420}>
                        {description}
                    </Text>
                ) : null}
            </Stack>
        </Stack>
    )
}
