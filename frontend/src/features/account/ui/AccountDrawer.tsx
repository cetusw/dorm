import { Drawer, FocusTrap } from '@mantine/core'

import type { CurrentUser } from '../../current-user/model/types'
import { AccountPanel } from './AccountPanel'
import classes from './AccountDrawer.module.css'

type Props = {
    currentUser: CurrentUser | null
    opened: boolean
    onClose: () => void
}

export function AccountDrawer({ currentUser, opened, onClose }: Props) {
    return (
        <Drawer
            opened={opened}
            onClose={onClose}
            position="bottom"
            hiddenFrom="md"
            size="100dvh"
            withCloseButton={false}
            classNames={{
                body: classes.content,
                content: classes.drawerContent,
            }}
        >
            <FocusTrap.InitialFocus />

            {currentUser ? (
                <AccountPanel
                    currentUser={currentUser}
                    variant="drawer"
                    onClose={onClose}
                />
            ) : null}
        </Drawer>
    )
}
