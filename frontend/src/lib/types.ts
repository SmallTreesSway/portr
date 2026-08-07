export const DirStatus = {
    NoneSelected: "NONSELECTED",
    Initializing: "INITIALIZING",
    ErrorOccured: "ERROROCCURED",
    Loaded: "LOADED",
} as const

export type DirStatus = typeof DirStatus[keyof typeof DirStatus];
