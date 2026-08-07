import { Box, Stack, Title } from '@mantine/core'
import { EmptyState } from '../../shared/ui/EmptyState'

type Props = {
    title: string
}

export function ResidentSectionPlaceholderPage({ title }: Props) {
    return (
        <Box px={{ base: 'md', md: 'xl' }} py="xl">
            <Stack gap="lg" maw={1240} mx="auto">
                <Title order={1}>{title}</Title>
                <EmptyState
                    title="Раздел недоступен"
                    description="Этот раздел пока не подключен."
                />
            </Stack>
        </Box>
    )
}
