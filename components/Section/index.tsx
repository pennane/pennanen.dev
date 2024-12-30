import { Stack, TStackProps } from '../Stack'

export const Section = ({ children, ...rest }: TStackProps) => {
	return (
		<section>
			<Stack {...rest}>{children}</Stack>
		</section>
	)
}
