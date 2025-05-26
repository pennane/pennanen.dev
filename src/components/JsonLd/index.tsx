import { Thing, WithContext } from 'schema-dts'

export const JsonLd = <T extends Thing>({
  object
}: {
  object: WithContext<T>
}) => {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(object) }}
    />
  )
}
