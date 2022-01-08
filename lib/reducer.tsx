export interface AppState {
    animations: {
        [animationKey: string]: boolean
    }
}

export enum AppActionType {
    ANIMATION_COMPLETE
}

export interface AppAction {
    type: AppActionType
    payload: string
}

export const initialState: AppState = {
    animations: {}
}

export const reducer = (state: AppState, action: AppAction): AppState => {
    switch (action.type) {
        case AppActionType.ANIMATION_COMPLETE:
            const _state = { ...state }
            _state.animations[action.payload] = true
            return _state

        default:
            return state
    }
}
