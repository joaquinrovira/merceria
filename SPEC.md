Let's plan a refactor of the codebase to make it modular, extendable and maintainable.
Remember to use go patterns and when in doubt prefer functional pattern over OOO style.
Let's build and iterate on a carefully planned specification.

The web API will have the following 3 core features:

- Can add rows to the liked spreadsheet based on a row-model that can (and will) evolve over time.
  For now, we expect each row to have the following format.

  Nº Pedido: A sequential number with the following format: YYYYMMDDnnnn where nnnn is a mononotinically increasing count
  Fecha: Row added date format (time.Time) serialized as YYYY/MM/DD
  Nombre: User input string
  Teléfono: User input string (phone number, but wont be parsed & validated)
  Nº Prendas: Integer, item count
  Notas: Arbitrary string
  Grupo: string, for now should follow the enum "Grupo 1", "Grupo 2", "Grupo 3" 
  Retraso: automatically calculated by excel based on formula "=IF(OR($C7="";$H7="");"";IF(TODAY()-INT($C7)>CHOOSE(VALUE(RIGHT($H7;1));7;10;15);"!!";""))"
  Estado: string enum TRUE or FALSE

The form should be a static html site added to static/
It can use tailwind for basic styling and alpineJS for form building.

The form, for now should only be able to handle a single row, but in the future a single form POST may add multiple rows of data.

- Can search rows. Should not be implemented right now but here is the idea.
  Build a SQLITE cache from spreadsheet data. Perhaps up to N entries (or up to X days: configurable).
  successfully adding rows to the spreadsheet should also be optimistically added to the cache.
  the cache should be validated every so often, if an id was recently added that was is not in the cache
  there was a 3rd party update and the cache should be rebuilt from scratch.
  sqlite with a reader connection and a writer connection.

- Can generate QR coded based on the Nº pedido. The idead is to have a link that will trigger integrations with other
  platforms: email, whatsapp business, etc. It must be open for extension.

Let's build and iterate on a carefully planned specification. 

We also want to take into account development mode. There should be a production server entrypoint (main) and a dev server entrypoint.