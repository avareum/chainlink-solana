import AccessController from './contracts/accessController'
import OCR2 from './contracts/ocr2'
import Token from './contracts/token'
import Validator from './contracts/validator'

export default [...AccessController, ...OCR2, ...Token, ...Validator]
