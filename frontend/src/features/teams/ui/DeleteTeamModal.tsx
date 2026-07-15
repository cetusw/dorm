import { EntityDeleteModal } from '../../../shared/ui/EntityDeleteModal'
import { deleteTeam } from '../api/teamsApi'
import type { TeamListItem } from '../model/types'

type Props = {
    opened: boolean
    team: TeamListItem | null
    onClose: () => void
    onDeleted: () => Promise<void> | void
}

export function DeleteTeamModal({ opened, team, onClose, onDeleted }: Props) {
    return (
        <EntityDeleteModal
            opened={opened}
            onClose={onClose}
            title="Удаление команды"
            entityLabel="команду"
            entityName={team?.name ?? null}
            errorMessage="Не удалось удалить команду"
            onConfirm={async () => {
                if (!team) {
                    return
                }

                await deleteTeam(team.id)
                await onDeleted()
            }}
        />
    )
}
