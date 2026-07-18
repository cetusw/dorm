import type { DutyAnalytics } from '../model/selectors'
import {
    buildMemberDutyProgressModel,
    buildReadonlyDutyProgressModel,
} from '../model/selectors'
import { MemberDutyProgressCard } from './MemberDutyProgressCard'

type Props = {
    analytics: DutyAnalytics
    isReadOnly: boolean
    targetValue: number
}

export function CurrentDutyAnalytics({
    analytics,
    isReadOnly,
    targetValue,
}: Props) {
    const progress = isReadOnly
        ? buildReadonlyDutyProgressModel(analytics)
        : buildMemberDutyProgressModel(analytics, targetValue)

    return <MemberDutyProgressCard progress={progress} />
}
