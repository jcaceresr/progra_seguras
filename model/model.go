package model

type Cliente struct {
	Nombre                string `json:"nombre" bson:"nombre"`
	Contrasena            string `json:"-" bson:"contrasena"`
	Fecha_nacimiento      string `json:"fecha_nacimiento" bson:"fecha_nacimiento"`
	Direccion             string `json:"direccion" bson:"direccion"`
	Numero_identificacion string `json:"numero_identificacion" bson:"numero_identificacion"`
	Email                 string `json:"email" bson:"email"`
	Telefono              string `json:"telefono" bson:"telefono"`
	Genero                string `json:"genero" bson:"genero"`
	Nacionalidad          string `json:"nacionalidad" bson:"nacionalidad"`
	Ocupacion             string `json:"ocupacion" bson:"ocupacion"`
}

type Billetera struct {
	Id          string `json:"id" bson:"_id,omitempty"` // El campo id debe ser exportado para que el driver de mongo pueda asignarle un valor
	Nro_cliente string `json:"nro_cliente" bson:"nro_cliente"`
	Saldo       string `json:"saldo" bson:"saldo"`
	Divisa      string `json:"divisa" bson:"divisa"`
	Nombre      string `json:"nombre" bson:"nombre"`
	Activo      bool   `json:"activo" bson:"activo"`
}

type Movimiento struct {
	Nro_cliente  string `json:"nro_cliente" bson:"nro_cliente"`
	Monto        string `json:"monto" bson:"monto"`
	Divisa       string `json:"divisa" bson:"divisa"`
	Tipo         string `json:"tipo" bson:"tipo"`
	Fecha_hora   string `json:"fecha_hora" bson:"fecha_hora"`
	ID_billetera string `json:"id_billetera" bson:"id_billetera"`
}

type Deposito struct {
	Nro_cliente string `json:"nro_cliente" bson:"nro_cliente"`
	Monto       string `json:"monto" bson:"monto"`
	Divisa      string `json:"divisa" bson:"divisa"`
	Tipo        string `json:"tipo" bson:"tipo"`
}

type Transferencia struct {
	Nro_cliente_origen  string `json:"nro_cliente_origen" bson:"nro_cliente_origen"`
	Nro_cliente_destino string `json:"nro_cliente_destino" bson:"nro_cliente_destino"`
	Monto               string `json:"monto" bson:"monto"`
	Divisa              string `json:"divisa" bson:"divisa"`
}

type GiroRequest struct {
	NroCliente string `json:"nro_cliente" bson:"nro_cliente"`
	Monto      string `json:"monto" bson:"monto"`
	Divisa     string `json:"divisa" bson:"divisa"`
}
