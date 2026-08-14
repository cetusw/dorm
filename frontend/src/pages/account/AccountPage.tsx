import type { CurrentUser } from '../../features/current-user/model/types'
import { AccountPanel } from '../../features/account/ui/AccountPanel'

type Props = {
    currentUser: CurrentUser
}

export function AccountPage({ currentUser }: Props) {
    return <AccountPanel currentUser={currentUser} variant="page" />
}
