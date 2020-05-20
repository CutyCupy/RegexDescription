class MyRegex extends RegExp {
    
    private static ESCAPABLE_CHARS = [] 

    private _pattern: string = "";


    private regex: IRegex = [];
    constructor(
        pattern: string
    ) {
        super(pattern);
        console.log(pattern);
        this.pattern = pattern;
    }

    public set pattern(pattern: string) {
        this._pattern = pattern;
        this.createRegex();
    }

    public createRegex() {
        this.regex = [];
        let compilerEscape = false;
        let regexEscape = false;
        let groupCreation: GroupCreation = undefined;;
        let groupStart = -1;
        let bracketCounter = 0;
        for(let i = 0; i < this._pattern.length; i++) {
            const char = this._pattern.charAt(i);
            if(regexEscape && compilerEscape) {
                this.addString(char);
                regexEscape = false;
                compilerEscape = false;
                continue;
            }
            switch(char) {
                case "\\":
                    if(groupCreation) {
                        break;
                    }
                    regexEscape = compilerEscape;
                    compilerEscape = true;
                    continue;
                case "(":
                    if(!groupCreation) {
                        groupCreation = 'group';
                        groupStart = i;
                    } 
                    if (groupCreation == 'group') {
                        bracketCounter++;
                    }
                    break;
                case ")":
                    if(groupCreation != 'group') {
                        break;
                    } 
                    bracketCounter--;
                    if(!bracketCounter) {
                        this.regex.push(new MyRegex(this._pattern.substring(groupStart + 1, i)));
                        groupCreation = undefined;;
                        groupStart = -1;
                    }
                    break;
                case "*":
                    if(groupCreation) {
                        break;
                    }
                    if(!compilerEscape && !regexEscape) {
                        this.regex.push({
                            min: 0,
                            max: Infinity,
                        });
                    }
                    break;
                case "+":
                    if(groupCreation) {
                        break;
                    }
                    if(!compilerEscape && !regexEscape) {
                        this.regex.push({
                            min: 1,
                            max: Infinity,
                        });
                    }
                    break;
                    case "?":
                    if(groupCreation) {
                        break;
                    }
                    if(!compilerEscape && !regexEscape) {
                        this.regex.push({
                            min: 0,
                            max: 1,
                        });
                    }
                case "{":
                    if(!groupCreation) {
                        groupCreation = 'quantifier';
                        groupStart = i;
                    } else if (groupCreation == 'quantifier') {
                        //TODO: Error?!
                        break;
                    }
                    break;
                case "}":
                    if(groupCreation != 'quantifier') {
                        break;
                    }
                    const quantifierText = this._pattern.substring(groupStart + 1, i);
                    const [min, max] = quantifierText.split(",");
                    this.regex.push({
                        min: Number(min.trim()),
                        max: max ? Number(max.trim()) : Number(min.trim()),
                    });
                    break;
                default:
                    if(groupCreation != undefined) {
                        break;
                    }
                    this.addString(char);
                    break;
            }
        }
    }

    public addString(toAdd: string) {
        let added = false;
        if(this.regex.length) {
            const last = this.regex[this.regex.length - 1];
            if (typeof last === "string") {
                this.regex[this.regex.length - 1] += toAdd;
                return;
            }
        }
        this.regex.push(toAdd)
    }

    public get description(): string {
        console.log(this.regex);
        return this._pattern.toString();
    }
}

type IRegex = IRegexPart[];

type IRegexPart = MyRegex | IOneOf[] | IQuantifier | string;

type GroupCreation = 'group' | 'quantifier' | 'oneOf' | undefined;

interface IOneOf {
    from: string;
    to: string;
}

interface IQuantifier {
    min: number;
    max: number;
}

new MyRegex("\\w").description;