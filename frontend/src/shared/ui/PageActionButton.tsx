import type { MouseEventHandler, ReactNode } from 'react'

import { PlusIcon } from '@phosphor-icons/react'
import { Button } from '@mantine/core'

type Props = {
    children: ReactNode
    disabled?: boolean
    onClick?: MouseEventHandler<HTMLButtonElement>
}

export function PageActionButton({ children, disabled = false, onClick }: Props) {
    return (
        <Button
            radius="md"
            h={42}
            leftSection={<PlusIcon size={18} />}
            disabled={disabled}
            onClick={onClick}
        >
            {children}
        </Button>
    )
}
