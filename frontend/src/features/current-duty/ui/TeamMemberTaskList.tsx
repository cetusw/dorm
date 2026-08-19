import { TaskGroups } from './TaskGroups'
import type { ResidentDutyTask } from '../model/types'

type Props = {
    tasks: ResidentDutyTask[]
}

export function TeamMemberTaskList({ tasks }: Props) {
    const noopTaskAction = async () => false

    return (
        <TaskGroups
            tasks={tasks}
            isReadOnly
            pendingTaskId={null}
            onTake={noopTaskAction}
            onReturn={noopTaskAction}
            onComplete={noopTaskAction}
            onOpen={noopTaskAction}
        />
    )
}
