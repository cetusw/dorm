import { Notification } from '@mantine/core'

import { HEADER_HEIGHT_PX } from './mobileStickyThreshold'

type Props = {
    color: string
    title: string
    message: string
    onClose: () => void
}

const FLOATING_NOTIFICATION_Z_INDEX = 40
const FLOATING_NOTIFICATION_TOP_OFFSET_PX = HEADER_HEIGHT_PX + 8

export function FloatingNotification({ color, title, message, onClose }: Props) {
    return (
        <>
            <div
                style={{
                    position: 'fixed',
                    top: FLOATING_NOTIFICATION_TOP_OFFSET_PX,
                    right: 16,
                    zIndex: FLOATING_NOTIFICATION_Z_INDEX,
                    width: 'min(420px, calc(100vw - 32px))',
                    animation: 'current-duty-notification-enter 220ms ease-out',
                }}
            >
                <Notification
                    withCloseButton
                    color={color}
                    title={title}
                    onClose={onClose}
                >
                    {message}
                </Notification>
            </div>
            <style>
                {`
                    @keyframes current-duty-notification-enter {
                        from {
                            opacity: 0;
                            transform: translateX(calc(100% + 16px));
                        }

                        to {
                            opacity: 1;
                            transform: translateX(0);
                        }
                    }
                `}
            </style>
        </>
    )
}
