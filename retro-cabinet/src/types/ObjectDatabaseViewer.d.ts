import { IStateErrorable, IStateLoadable, ICheckpoint } from './index'

/*
 * Redux State
 */
export interface IState extends IStateErrorable, IStateLoadable {
    readonly selectedHeadRefHash: string
    readonly checkpoints: ICheckpoint[]
} 

/*
 * Component Props
 */
interface IPropsFromState {
    checkpoints: ICheckpoint[]
    isLoading: boolean
    selectedHeadRefHash: string
}

// tslint:disable-next-line:no-empty-interface
interface IPropsFromDispatch {}

// tslint:disable-next-line:no-empty-interface
export type IProps = IPropsFromState | IPropsFromDispatch;