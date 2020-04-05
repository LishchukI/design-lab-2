package gomodule

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"path"
)

var (
	pctx = blueprint.NewPackageContext("github.com/KPI-Labs/design-lab-2/build/gomodule")

	goBuild = pctx.StaticRule("binaryBuild", blueprint.RuleParams{
		Command:     "cd $workDir && go build -o $outputPath $pkg",
		Description: "build go command $pkg",
	}, "workDir", "outputPath", "pkg")

	goTest = pctx.StaticRule("test", blueprint.RuleParams{
		Command:     "cd $workDir && go test -v $testPkg > $outputPath",
		Description: "test $testPkg",
	}, "workDir", "outputPath", "testPkg")
)

type testedBinaryModule struct {
	blueprint.SimpleName

	properties struct {
		Pkg string
		Srcs []string
		TestPkg string
		TestSrcs []string
	}
}

func convertPatternsIntoPaths(ctx blueprint.ModuleContext, patterns []string, excludePatterns []string) []string {
	var paths []string
	for _, src := range patterns {
		if matches, err := ctx.GlobWithDeps(src, excludePatterns); err == nil {
			paths = append(paths, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			return nil
		}
	}
	return paths
}

func (gb *testedBinaryModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	pathToBin := path.Join(config.BaseOutputDir, "bin", name)
	pathToReports := path.Join(config.BaseOutputDir, "reports", name, "test.txt")

	inputs := convertPatternsIntoPaths(ctx, gb.properties.Srcs, gb.properties.TestSrcs)
	testInputs := convertPatternsIntoPaths(ctx, gb.properties.TestSrcs, []string{})

	if inputs != nil {
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Build %s as Go binary", name),
			Rule:        goBuild,
			Outputs:     []string{pathToBin},
			Implicits:   inputs,
			Args: map[string]string{
				"outputPath": pathToBin,
				"workDir":    ctx.ModuleDir(),
				"pkg":        gb.properties.Pkg,
			},
		})
	}

	if testInputs != nil {
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Test my module"),
			Rule:        goTest,
			Outputs:     []string{pathToReports},
			Implicits:   append(testInputs, inputs...),
			Args: map[string]string{
				"outputPath": pathToReports,
				"workDir":    ctx.ModuleDir(),
				"testPkg":    gb.properties.TestPkg,
			},
		})
	}
}

func SimpleBinFactory() (blueprint.Module, []interface{}) {
	mType := &testedBinaryModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
