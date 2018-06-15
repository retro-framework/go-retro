import { IStateErrorable, IStateLoadable, IRefHash } from './index'

/*
 * Redux State
 */ 
export interface IState extends IStateErrorable, IStateLoadable {
    readonly refs: IRefHash[]
    readonly selectedHash: string
}

/*
 * Component Props
 */
interface IPropsFromState extends IStateErrorable, IStateLoadable {
    refs: IRefHash[]
    selectedHash: string
}

interface IPropsFromDispatch {
    handleChangedSelectedHeadRefHash: (_: string) => void
}

export type IProps = IPropsFromState | IPropsFromDispatch