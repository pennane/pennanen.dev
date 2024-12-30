export type TProject = {
	id: string
	url: string | null
	date: number | null
	ignoreDate: boolean
	ignoreInListing: boolean
	name: string
	description: string | null
	pretext: string | null
	icon: string | null
	images: string[]
	largeImage: string | null
	github: string | null
	notAuthor: boolean | null
}

export type TProjectConfig = {
	url: string | null
	date: string | number | null
	name: {
		en: string | null
		fi: string | null
	}
	description: {
		en: string | null
		fi: string | null
	}
	icon: string | null
	ignoreInListing?: boolean
	ignoreDate?: boolean
	largeImage?: string
	images?: string[]
	pretext?: string
	github?: string
	notAuthor?: boolean
}
